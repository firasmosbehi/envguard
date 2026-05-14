package validator

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/envguard/envguard/internal/schema"
)

// Validate checks the given env vars against the schema.
// If strict is true, warnings are generated for keys present in envVars but not in the schema.
// envName is the current environment (e.g. "production", "development") for environment-specific rules.
func Validate(s *schema.Schema, envVars map[string]string, strict bool, envName string) *Result {
	return ValidateParallel(s, envVars, strict, envName, false)
}

// ValidateParallel checks the given env vars with optional parallel validation.
func ValidateParallel(s *schema.Schema, envVars map[string]string, strict bool, envName string, parallel bool) *Result {
	result := NewResult()

	if parallel && len(s.Env) > 1 {
		var mu sync.Mutex
		var wg sync.WaitGroup

		for name, variable := range s.Env {
			wg.Add(1)
			go func(n string, v *schema.Variable) {
				defer wg.Done()
				rawValue, exists := envVars[n]
				subResult := NewResult()
				validateVariable(subResult, n, v, rawValue, exists, envName, envVars)
				mu.Lock()
				mergeResults(result, subResult)
				mu.Unlock()
			}(name, variable)
		}
		wg.Wait()
	} else {
		for name, variable := range s.Env {
			rawValue, exists := envVars[name]
			validateVariable(result, name, variable, rawValue, exists, envName, envVars)
		}
	}

	// Strict mode: warn about unknown keys in .env
	if strict {
		for name := range envVars {
			if _, defined := s.Env[name]; !defined {
				result.AddWarning(name, "strict", "variable is not defined in schema")
			}
		}
	}

	return result
}

func mergeResults(dst, src *Result) {
	if !src.Valid {
		dst.Valid = false
	}
	dst.Errors = append(dst.Errors, src.Errors...)
	dst.Warnings = append(dst.Warnings, src.Warnings...)
}

func severityFor(variable *schema.Variable) Severity {
	switch variable.Severity {
	case "warn":
		return SeverityWarn
	case "info":
		return SeverityInfo
	default:
		return SeverityError
	}
}

func addValidationError(result *Result, name, rule, message string, variable *schema.Variable) {
	result.AddErrorWithSeverity(name, rule, message, severityFor(variable))
}

func validateVariable(result *Result, name string, variable *schema.Variable, rawValue string, exists bool, envName string, envVars map[string]string) {
	// 0. Warn if deprecated and present
	if variable.Deprecated != "" && exists {
		result.AddWarning(name, "deprecated", fmt.Sprintf("variable is deprecated: %s", variable.Deprecated))
	}

	// 1. Check devOnly: ignore in non-dev environments, required in dev
	required := variable.Required
	if variable.DevOnly {
		if envName != "" && envName != "development" && envName != "dev" {
			return
		}
		required = true
	}

	// 2. Check requiredIn
	if len(variable.RequiredIn) > 0 {
		if envName != "" {
			required = false
			for _, env := range variable.RequiredIn {
				if strings.EqualFold(env, envName) {
					required = true
					break
				}
			}
		}
	}

	// 3. Check conditional dependency (dependsOn + when)
	if variable.DependsOn != "" && variable.When != "" {
		depValue, depExists := envVars[variable.DependsOn]
		if depExists && strings.EqualFold(strings.TrimSpace(depValue), variable.When) {
			required = true
		}
	}

	// 4. Check required
	if required {
		if !exists || strings.TrimSpace(rawValue) == "" {
			msg := "variable is missing or empty"
			if variable.Message != "" {
				msg = variable.Message
			}
			addValidationError(result, name, "required", msg, variable)
			return
		}
	}

	// 5. Check allowEmpty
	if variable.AllowEmpty != nil && !*variable.AllowEmpty {
		if exists && strings.TrimSpace(rawValue) == "" {
			addValidationError(result, name, "allowEmpty", customMessage(variable, "value cannot be empty"), variable)
			return
		}
	}

	// 5.5 Check notEmpty for arrays
	if variable.NotEmpty != nil && *variable.NotEmpty && variable.Type == schema.TypeArray {
		if exists && rawValue == "" {
			addValidationError(result, name, "notEmpty", customMessage(variable, "array cannot be empty"), variable)
			return
		}
	}

	// 6. Apply default if missing
	if !exists || rawValue == "" {
		if variable.Default != nil {
			rawValue = defaultToString(variable.Default)
		} else {
			// Optional and no default: skip further validation
			return
		}
	}

	// 6.5 Apply transform
	rawValue = applyTransform(rawValue, variable.Transform)

	// 7. Coerce to type and validate
	switch variable.Type {
	case schema.TypeString:
		validateString(result, name, variable, rawValue)
	case schema.TypeInteger:
		validateInteger(result, name, variable, rawValue)
	case schema.TypeFloat:
		validateFloat(result, name, variable, rawValue)
	case schema.TypeBoolean:
		validateBoolean(result, name, variable, rawValue)
	case schema.TypeArray:
		validateArray(result, name, variable, rawValue)
	}
}

func applyTransform(value, transform string) string {
	switch transform {
	case "lowercase":
		return strings.ToLower(value)
	case "uppercase":
		return strings.ToUpper(value)
	case "trim":
		return strings.TrimSpace(value)
	default:
		return value
	}
}

func customMessage(variable *schema.Variable, defaultMsg string) string {
	if variable.Message != "" {
		return variable.Message
	}
	return defaultMsg
}

func validateString(result *Result, name string, variable *schema.Variable, rawValue string) {
	value, err := coerceString(rawValue)
	if err != nil {
		addValidationError(result, name, "type", customMessage(variable, err.Error()), variable)
		return
	}

	if variable.MinLength != nil && len(value) < *variable.MinLength {
		addValidationError(result, name, "minLength", customMessage(variable, fmt.Sprintf("value has length %d, expected at least %d", len(value), *variable.MinLength)), variable)
	}

	if variable.MaxLength != nil && len(value) > *variable.MaxLength {
		addValidationError(result, name, "maxLength", customMessage(variable, fmt.Sprintf("value has length %d, expected at most %d", len(value), *variable.MaxLength)), variable)
	}

	if variable.Prefix != "" && !strings.HasPrefix(value, variable.Prefix) {
		addValidationError(result, name, "prefix", customMessage(variable, fmt.Sprintf("value %q does not start with prefix %q", value, variable.Prefix)), variable)
	}

	if variable.Suffix != "" && !strings.HasSuffix(value, variable.Suffix) {
		addValidationError(result, name, "suffix", customMessage(variable, fmt.Sprintf("value %q does not end with suffix %q", value, variable.Suffix)), variable)
	}

	if variable.Format != "" {
		if err := validateFormat(value, variable.Format); err != nil {
			addValidationError(result, name, "format", customMessage(variable, err.Error()), variable)
		}
	}

	if len(variable.Disallow) > 0 {
		for _, disallowed := range variable.Disallow {
			if value == disallowed {
				addValidationError(result, name, "disallow", customMessage(variable, fmt.Sprintf("value %q is not allowed", value)), variable)
				break
			}
		}
	}

	if variable.Pattern != "" {
		re, err := regexCache.Compile(variable.Pattern)
		if err != nil {
			addValidationError(result, name, "pattern", customMessage(variable, fmt.Sprintf("invalid regex pattern: %v", err)), variable)
			return
		}
		if !re.MatchString(value) {
			addValidationError(result, name, "pattern", customMessage(variable, fmt.Sprintf("value %q does not match pattern %q", value, variable.Pattern)), variable)
		}
	}

	if variable.Enum != nil {
		if len(variable.Enum) == 0 {
			addValidationError(result, name, "enum", customMessage(variable, "no values are allowed (enum is empty)"), variable)
		} else if !stringInSlice(value, variable.Enum) {
			addValidationError(result, name, "enum", customMessage(variable, fmt.Sprintf("value %q is not one of allowed values", value)), variable)
		}
	}
}

func validateInteger(result *Result, name string, variable *schema.Variable, rawValue string) {
	value, err := coerceInteger(rawValue)
	if err != nil {
		addValidationError(result, name, "type", customMessage(variable, err.Error()), variable)
		return
	}

	if variable.Min != nil {
		if minVal, ok := schemaToInt64(variable.Min); ok && value < minVal {
			addValidationError(result, name, "min", customMessage(variable, fmt.Sprintf("value %d is less than minimum %d", value, minVal)), variable)
		}
	}

	if variable.Max != nil {
		if maxVal, ok := schemaToInt64(variable.Max); ok && value > maxVal {
			addValidationError(result, name, "max", customMessage(variable, fmt.Sprintf("value %d is greater than maximum %d", value, maxVal)), variable)
		}
	}

	if variable.MultipleOf != nil {
		if mult, ok := schemaToFloat64(variable.MultipleOf); ok && mult != 0 {
			if float64(value)/mult != float64(int64(float64(value)/mult)) {
				addValidationError(result, name, "multipleOf", customMessage(variable, fmt.Sprintf("value %d is not a multiple of %v", value, variable.MultipleOf)), variable)
			}
		}
	}

	if variable.Enum != nil {
		if len(variable.Enum) == 0 {
			addValidationError(result, name, "enum", customMessage(variable, "no values are allowed (enum is empty)"), variable)
		} else if !int64InSlice(value, variable.Enum) {
			addValidationError(result, name, "enum", customMessage(variable, fmt.Sprintf("value %d is not one of allowed values", value)), variable)
		}
	}
}

func validateFloat(result *Result, name string, variable *schema.Variable, rawValue string) {
	value, err := coerceFloat(rawValue)
	if err != nil {
		addValidationError(result, name, "type", customMessage(variable, err.Error()), variable)
		return
	}

	if variable.Min != nil {
		if minVal, ok := schemaToFloat64(variable.Min); ok && value < minVal {
			addValidationError(result, name, "min", customMessage(variable, fmt.Sprintf("value %g is less than minimum %g", value, minVal)), variable)
		}
	}

	if variable.Max != nil {
		if maxVal, ok := schemaToFloat64(variable.Max); ok && value > maxVal {
			addValidationError(result, name, "max", customMessage(variable, fmt.Sprintf("value %g is greater than maximum %g", value, maxVal)), variable)
		}
	}

	if variable.MultipleOf != nil {
		if mult, ok := schemaToFloat64(variable.MultipleOf); ok && mult != 0 {
			if float64(int64(value/mult))*mult != value {
				addValidationError(result, name, "multipleOf", customMessage(variable, fmt.Sprintf("value %g is not a multiple of %v", value, variable.MultipleOf)), variable)
			}
		}
	}

	if variable.Enum != nil {
		if len(variable.Enum) == 0 {
			addValidationError(result, name, "enum", customMessage(variable, "no values are allowed (enum is empty)"), variable)
		} else if !float64InSlice(value, variable.Enum) {
			addValidationError(result, name, "enum", customMessage(variable, fmt.Sprintf("value %g is not one of allowed values", value)), variable)
		}
	}
}

func validateBoolean(result *Result, name string, variable *schema.Variable, rawValue string) {
	_, err := coerceBoolean(rawValue)
	if err != nil {
		addValidationError(result, name, "type", customMessage(variable, err.Error()), variable)
	}
}

func validateArray(result *Result, name string, variable *schema.Variable, rawValue string) {
	if rawValue == "" {
		addValidationError(result, name, "type", customMessage(variable, "expected array, got empty string"), variable)
		return
	}

	items := strings.Split(rawValue, variable.Separator)
	if variable.MinLength != nil && len(items) < *variable.MinLength {
		addValidationError(result, name, "minLength", customMessage(variable, fmt.Sprintf("array has %d items, expected at least %d", len(items), *variable.MinLength)), variable)
	}
	if variable.MaxLength != nil && len(items) > *variable.MaxLength {
		addValidationError(result, name, "maxLength", customMessage(variable, fmt.Sprintf("array has %d items, expected at most %d", len(items), *variable.MaxLength)), variable)
	}

	if variable.NotEmpty != nil && *variable.NotEmpty && len(items) == 0 {
		addValidationError(result, name, "notEmpty", customMessage(variable, "array cannot be empty"), variable)
	}

	if variable.UniqueItems {
		seen := make(map[string]bool)
		for _, item := range items {
			item = strings.TrimSpace(item)
			if seen[item] {
				addValidationError(result, name, "uniqueItems", customMessage(variable, fmt.Sprintf("duplicate item %q", item)), variable)
				break
			}
			seen[item] = true
		}
	}

	if variable.ItemPattern != "" {
		re, err := regexCache.Compile(variable.ItemPattern)
		if err != nil {
			addValidationError(result, name, "itemPattern", customMessage(variable, fmt.Sprintf("invalid regex pattern: %v", err)), variable)
		} else {
			for _, item := range items {
				item = strings.TrimSpace(item)
				if !re.MatchString(item) {
					addValidationError(result, name, "itemPattern", customMessage(variable, fmt.Sprintf("item %q does not match pattern %q", item, variable.ItemPattern)), variable)
				}
			}
		}
	}

	if variable.ItemType != "" {
		for _, item := range items {
			item = strings.TrimSpace(item)
			switch variable.ItemType {
			case schema.TypeInteger:
				if _, err := coerceInteger(item); err != nil {
					addValidationError(result, name, "itemType", customMessage(variable, fmt.Sprintf("item %q is not an integer: %v", item, err)), variable)
				}
			case schema.TypeFloat:
				if _, err := coerceFloat(item); err != nil {
					addValidationError(result, name, "itemType", customMessage(variable, fmt.Sprintf("item %q is not a float: %v", item, err)), variable)
				}
			case schema.TypeBoolean:
				if _, err := coerceBoolean(item); err != nil {
					addValidationError(result, name, "itemType", customMessage(variable, fmt.Sprintf("item %q is not a boolean: %v", item, err)), variable)
				}
			}
		}
	}

	if len(variable.Enum) > 0 {
		for _, item := range items {
			item = strings.TrimSpace(item)
			if !stringInSlice(item, variable.Enum) {
				addValidationError(result, name, "enum", customMessage(variable, fmt.Sprintf("item %q is not one of allowed values", item)), variable)
			}
		}
	}

	if variable.Contains != "" {
		found := false
		for _, item := range items {
			if strings.TrimSpace(item) == variable.Contains {
				found = true
				break
			}
		}
		if !found {
			addValidationError(result, name, "contains", customMessage(variable, fmt.Sprintf("array does not contain %q", variable.Contains)), variable)
		}
	}
}

func validateFormat(value, format string) error {
	switch format {
	case "email":
		_, err := mail.ParseAddress(value)
		if err != nil {
			return fmt.Errorf("value %q is not a valid email address", value)
		}
		// Ensure it's just an email, not "Name <email>"
		if strings.Contains(value, "<") || strings.Contains(value, ">") {
			return fmt.Errorf("value %q is not a valid email address", value)
		}
	case "url":
		u, err := url.Parse(value)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("value %q is not a valid URL", value)
		}
	case "uuid":
		uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
		if !uuidRegex.MatchString(value) {
			return fmt.Errorf("value %q is not a valid UUID", value)
		}
	case "base64":
		if _, err := base64.StdEncoding.DecodeString(value); err != nil {
			return fmt.Errorf("value %q is not valid base64", value)
		}
	case "ip":
		if net.ParseIP(value) == nil {
			return fmt.Errorf("value %q is not a valid IP address", value)
		}
	case "port":
		port, err := strconv.Atoi(value)
		if err != nil || port < 1 || port > 65535 {
			return fmt.Errorf("value %q is not a valid port (1-65535)", value)
		}
	case "json":
		var js json.RawMessage
		if err := json.Unmarshal([]byte(value), &js); err != nil {
			return fmt.Errorf("value %q is not valid JSON", value)
		}
	case "duration":
		if _, err := time.ParseDuration(value); err != nil {
			return fmt.Errorf("value %q is not a valid duration", value)
		}
	case "semver":
		semverRegex := regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
		if !semverRegex.MatchString(value) {
			return fmt.Errorf("value %q is not a valid semantic version", value)
		}
	case "hostname":
		if len(value) > 253 || value == "" || strings.HasPrefix(value, "-") || strings.HasSuffix(value, "-") {
			return fmt.Errorf("value %q is not a valid hostname", value)
		}
		hostnameRegex := regexp.MustCompile(`^[A-Za-z0-9-]{1,63}(\.[A-Za-z0-9-]{1,63})*\.?$`)
		if !hostnameRegex.MatchString(value) {
			return fmt.Errorf("value %q is not a valid hostname", value)
		}
	case "hex":
		hexRegex := regexp.MustCompile(`^(0x)?[0-9a-fA-F]+$`)
		if !hexRegex.MatchString(value) {
			return fmt.Errorf("value %q is not valid hexadecimal", value)
		}
	case "cron":
		cronRegex := regexp.MustCompile(`^(@(annually|yearly|monthly|weekly|daily|hourly|reboot))|((((\d+,)+\d+|([\d*]+(/\d+)?)|\d+|\*)\s?){5,7})$`)
		if !cronRegex.MatchString(value) {
			return fmt.Errorf("value %q is not a valid cron expression", value)
		}
	case "datetime":
		// ISO 8601 timestamp
		if _, err := time.Parse(time.RFC3339, value); err != nil {
			return fmt.Errorf("value %q is not a valid ISO 8601 datetime (RFC3339)", value)
		}
	case "date":
		if _, err := time.Parse("2006-01-02", value); err != nil {
			return fmt.Errorf("value %q is not a valid date (YYYY-MM-DD)", value)
		}
	case "time":
		if _, err := time.Parse("15:04:05", value); err != nil {
			return fmt.Errorf("value %q is not a valid time (HH:MM:SS)", value)
		}
	case "timezone":
		if _, err := time.LoadLocation(value); err != nil {
			return fmt.Errorf("value %q is not a valid IANA timezone", value)
		}
	case "color":
		hexRegex := regexp.MustCompile(`^#([0-9a-fA-F]{3}){1,2}$`)
		rgbRegex := regexp.MustCompile(`^rgb\(\s*\d{1,3}\s*,\s*\d{1,3}\s*,\s*\d{1,3}\s*\)$`)
		rgbaRegex := regexp.MustCompile(`^rgba\(\s*\d{1,3}\s*,\s*\d{1,3}\s*,\s*\d{1,3}\s*,\s*(0|1|0?\.\d+)\s*\)$`)
		hslRegex := regexp.MustCompile(`^hsl\(\s*\d{1,3}\s*,\s*\d{1,3}%?\s*,\s*\d{1,3}%?\s*\)$`)
		if !hexRegex.MatchString(value) && !rgbRegex.MatchString(value) && !rgbaRegex.MatchString(value) && !hslRegex.MatchString(value) {
			return fmt.Errorf("value %q is not a valid color (hex, rgb, rgba, hsl)", value)
		}
	case "slug":
		slugRegex := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
		if !slugRegex.MatchString(value) {
			return fmt.Errorf("value %q is not a valid slug", value)
		}
	case "filepath":
		if value == "" {
			return fmt.Errorf("value is not a valid file path")
		}
		if strings.Contains(value, "\x00") {
			return fmt.Errorf("value %q contains null bytes", value)
		}
	case "directory":
		if value == "" {
			return fmt.Errorf("value is not a valid directory path")
		}
		if strings.Contains(value, "\x00") {
			return fmt.Errorf("value %q contains null bytes", value)
		}
	case "locale":
		localeRegex := regexp.MustCompile(`^[a-zA-Z]{2,3}(-[a-zA-Z]{2,4})?(-[a-zA-Z0-9]+)*$`)
		if !localeRegex.MatchString(value) {
			return fmt.Errorf("value %q is not a valid BCP 47 locale", value)
		}
	case "jwt":
		jwtRegex := regexp.MustCompile(`^[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]*$`)
		if !jwtRegex.MatchString(value) {
			return fmt.Errorf("value %q is not a valid JWT format", value)
		}
	case "mongodb-uri":
		if !strings.HasPrefix(value, "mongodb://") && !strings.HasPrefix(value, "mongodb+srv://") {
			return fmt.Errorf("value %q is not a valid MongoDB URI", value)
		}
		u, err := url.Parse(value)
		if err != nil || u.Host == "" {
			return fmt.Errorf("value %q is not a valid MongoDB URI", value)
		}
	case "redis-uri":
		if !strings.HasPrefix(value, "redis://") && !strings.HasPrefix(value, "rediss://") {
			return fmt.Errorf("value %q is not a valid Redis URI", value)
		}
		u, err := url.Parse(value)
		if err != nil || u.Host == "" {
			return fmt.Errorf("value %q is not a valid Redis URI", value)
		}
	}
	return nil
}

func schemaToInt64(value any) (int64, bool) {
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case uint:
		return int64(v), true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		return int64(v), true
	case float64:
		return int64(v), true
	case float32:
		return int64(v), true
	default:
		return 0, false
	}
}

func schemaToFloat64(value any) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}

func defaultToString(def any) string {
	switch v := def.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func stringInSlice(s string, slice []any) bool {
	for _, v := range slice {
		if vs, ok := v.(string); ok && vs == s {
			return true
		}
	}
	return false
}

func int64InSlice(n int64, slice []any) bool {
	for _, v := range slice {
		switch vt := v.(type) {
		case int:
			if int64(vt) == n {
				return true
			}
		case int64:
			if vt == n {
				return true
			}
		case float64:
			if vt == float64(n) {
				return true
			}
		}
	}
	return false
}

func float64InSlice(f float64, slice []any) bool {
	for _, v := range slice {
		switch vt := v.(type) {
		case int:
			if float64(vt) == f {
				return true
			}
		case int64:
			if float64(vt) == f {
				return true
			}
		case float64:
			if vt == f {
				return true
			}
		}
	}
	return false
}
