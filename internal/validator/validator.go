package validator

import (
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/envguard/envguard/internal/schema"
)

// Validate checks the given env vars against the schema.
// If strict is true, warnings are generated for keys present in envVars but not in the schema.
// envName is the current environment (e.g. "production", "development") for environment-specific rules.
func Validate(s *schema.Schema, envVars map[string]string, strict bool, envName string) *Result {
	result := NewResult()

	// Validate each variable defined in the schema
	for name, variable := range s.Env {
		rawValue, exists := envVars[name]
		validateVariable(result, name, variable, rawValue, exists, envName)
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

func validateVariable(result *Result, name string, variable *schema.Variable, rawValue string, exists bool, envName string) {
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

	// 3. Check required
	if required {
		if !exists || strings.TrimSpace(rawValue) == "" {
			msg := "variable is missing or empty"
			if variable.Message != "" {
				msg = variable.Message
			}
			result.AddError(name, "required", msg)
			return
		}
	}

	// 4. Apply default if missing
	if !exists || rawValue == "" {
		if variable.Default != nil {
			rawValue = defaultToString(variable.Default)
			exists = true
		} else {
			// Optional and no default: skip further validation
			return
		}
	}

	// 5. Coerce to type and validate
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

func customMessage(variable *schema.Variable, defaultMsg string) string {
	if variable.Message != "" {
		return variable.Message
	}
	return defaultMsg
}

func validateString(result *Result, name string, variable *schema.Variable, rawValue string) {
	value, err := coerceString(rawValue)
	if err != nil {
		result.AddError(name, "type", customMessage(variable, err.Error()))
		return
	}

	if variable.MinLength != nil && len(value) < *variable.MinLength {
		result.AddError(name, "minLength", customMessage(variable, fmt.Sprintf("value has length %d, expected at least %d", len(value), *variable.MinLength)))
	}

	if variable.MaxLength != nil && len(value) > *variable.MaxLength {
		result.AddError(name, "maxLength", customMessage(variable, fmt.Sprintf("value has length %d, expected at most %d", len(value), *variable.MaxLength)))
	}

	if variable.Format != "" {
		if err := validateFormat(value, variable.Format); err != nil {
			result.AddError(name, "format", customMessage(variable, err.Error()))
		}
	}

	if len(variable.Disallow) > 0 {
		for _, disallowed := range variable.Disallow {
			if value == disallowed {
				result.AddError(name, "disallow", customMessage(variable, fmt.Sprintf("value %q is not allowed", value)))
				break
			}
		}
	}

	if variable.Pattern != "" {
		re, err := regexp.Compile(variable.Pattern)
		if err != nil {
			result.AddError(name, "pattern", customMessage(variable, fmt.Sprintf("invalid regex pattern: %v", err)))
			return
		}
		if !re.MatchString(value) {
			result.AddError(name, "pattern", customMessage(variable, fmt.Sprintf("value %q does not match pattern %q", value, variable.Pattern)))
		}
	}

	if variable.Enum != nil {
		if len(variable.Enum) == 0 {
			result.AddError(name, "enum", customMessage(variable, "no values are allowed (enum is empty)"))
		} else if !stringInSlice(value, variable.Enum) {
			result.AddError(name, "enum", customMessage(variable, fmt.Sprintf("value %q is not one of allowed values", value)))
		}
	}
}

func validateInteger(result *Result, name string, variable *schema.Variable, rawValue string) {
	value, err := coerceInteger(rawValue)
	if err != nil {
		result.AddError(name, "type", customMessage(variable, err.Error()))
		return
	}

	if variable.Min != nil {
		if minVal, ok := schemaToInt64(variable.Min); ok && value < minVal {
			result.AddError(name, "min", customMessage(variable, fmt.Sprintf("value %d is less than minimum %d", value, minVal)))
		}
	}

	if variable.Max != nil {
		if maxVal, ok := schemaToInt64(variable.Max); ok && value > maxVal {
			result.AddError(name, "max", customMessage(variable, fmt.Sprintf("value %d is greater than maximum %d", value, maxVal)))
		}
	}

	if variable.Enum != nil {
		if len(variable.Enum) == 0 {
			result.AddError(name, "enum", customMessage(variable, "no values are allowed (enum is empty)"))
		} else if !int64InSlice(value, variable.Enum) {
			result.AddError(name, "enum", customMessage(variable, fmt.Sprintf("value %d is not one of allowed values", value)))
		}
	}
}

func validateFloat(result *Result, name string, variable *schema.Variable, rawValue string) {
	value, err := coerceFloat(rawValue)
	if err != nil {
		result.AddError(name, "type", customMessage(variable, err.Error()))
		return
	}

	if variable.Min != nil {
		if minVal, ok := schemaToFloat64(variable.Min); ok && value < minVal {
			result.AddError(name, "min", customMessage(variable, fmt.Sprintf("value %g is less than minimum %g", value, minVal)))
		}
	}

	if variable.Max != nil {
		if maxVal, ok := schemaToFloat64(variable.Max); ok && value > maxVal {
			result.AddError(name, "max", customMessage(variable, fmt.Sprintf("value %g is greater than maximum %g", value, maxVal)))
		}
	}

	if variable.Enum != nil {
		if len(variable.Enum) == 0 {
			result.AddError(name, "enum", customMessage(variable, "no values are allowed (enum is empty)"))
		} else if !float64InSlice(value, variable.Enum) {
			result.AddError(name, "enum", customMessage(variable, fmt.Sprintf("value %g is not one of allowed values", value)))
		}
	}
}

func validateBoolean(result *Result, name string, variable *schema.Variable, rawValue string) {
	_, err := coerceBoolean(rawValue)
	if err != nil {
		result.AddError(name, "type", customMessage(variable, err.Error()))
	}
}

func validateArray(result *Result, name string, variable *schema.Variable, rawValue string) {
	if rawValue == "" {
		result.AddError(name, "type", customMessage(variable, "expected array, got empty string"))
		return
	}

	items := strings.Split(rawValue, variable.Separator)
	if variable.MinLength != nil && len(items) < *variable.MinLength {
		result.AddError(name, "minLength", customMessage(variable, fmt.Sprintf("array has %d items, expected at least %d", len(items), *variable.MinLength)))
	}
	if variable.MaxLength != nil && len(items) > *variable.MaxLength {
		result.AddError(name, "maxLength", customMessage(variable, fmt.Sprintf("array has %d items, expected at most %d", len(items), *variable.MaxLength)))
	}

	if len(variable.Enum) > 0 {
		for _, item := range items {
			item = strings.TrimSpace(item)
			if !stringInSlice(item, variable.Enum) {
				result.AddError(name, "enum", customMessage(variable, fmt.Sprintf("item %q is not one of allowed values", item)))
			}
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
