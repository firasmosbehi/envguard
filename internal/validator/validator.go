package validator

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/envguard/envguard/internal/schema"
)

// Validate checks the given env vars against the schema.
// If strict is true, warnings are generated for keys present in envVars but not in the schema.
func Validate(s *schema.Schema, envVars map[string]string, strict bool) *Result {
	result := NewResult()

	// Validate each variable defined in the schema
	for name, variable := range s.Env {
		rawValue, exists := envVars[name]
		validateVariable(result, name, variable, rawValue, exists)
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

func validateVariable(result *Result, name string, variable *schema.Variable, rawValue string, exists bool) {
	// 1. Check required
	if variable.Required {
		if !exists || strings.TrimSpace(rawValue) == "" {
			result.AddError(name, "required", "variable is missing or empty")
			return
		}
	}

	// 2. Apply default if missing
	if !exists || rawValue == "" {
		if variable.Default != nil {
			rawValue = defaultToString(variable.Default)
			exists = true
		} else {
			// Optional and no default: skip further validation
			return
		}
	}

	// 3. Coerce to type and validate
	switch variable.Type {
	case schema.TypeString:
		validateString(result, name, variable, rawValue)
	case schema.TypeInteger:
		validateInteger(result, name, variable, rawValue)
	case schema.TypeFloat:
		validateFloat(result, name, variable, rawValue)
	case schema.TypeBoolean:
		validateBoolean(result, name, rawValue)
	}
}

func validateString(result *Result, name string, variable *schema.Variable, rawValue string) {
	value, err := coerceString(rawValue)
	if err != nil {
		result.AddError(name, "type", err.Error())
		return
	}

	if variable.Pattern != "" {
		re, err := regexp.Compile(variable.Pattern)
		if err != nil {
			result.AddError(name, "pattern", fmt.Sprintf("invalid regex pattern: %v", err))
			return
		}
		if !re.MatchString(value) {
			result.AddError(name, "pattern", fmt.Sprintf("value %q does not match pattern %q", value, variable.Pattern))
		}
	}

	if variable.Enum != nil {
		if len(variable.Enum) == 0 {
			result.AddError(name, "enum", "no values are allowed (enum is empty)")
		} else if !stringInSlice(value, variable.Enum) {
			result.AddError(name, "enum", fmt.Sprintf("value %q is not one of allowed values", value))
		}
	}
}

func validateInteger(result *Result, name string, variable *schema.Variable, rawValue string) {
	value, err := coerceInteger(rawValue)
	if err != nil {
		result.AddError(name, "type", err.Error())
		return
	}

	if variable.Enum != nil {
		if len(variable.Enum) == 0 {
			result.AddError(name, "enum", "no values are allowed (enum is empty)")
		} else if !int64InSlice(value, variable.Enum) {
			result.AddError(name, "enum", fmt.Sprintf("value %d is not one of allowed values", value))
		}
	}
}

func validateFloat(result *Result, name string, variable *schema.Variable, rawValue string) {
	value, err := coerceFloat(rawValue)
	if err != nil {
		result.AddError(name, "type", err.Error())
		return
	}

	if variable.Enum != nil {
		if len(variable.Enum) == 0 {
			result.AddError(name, "enum", "no values are allowed (enum is empty)")
		} else if !float64InSlice(value, variable.Enum) {
			result.AddError(name, "enum", fmt.Sprintf("value %f is not one of allowed values", value))
		}
	}
}

func validateBoolean(result *Result, name string, rawValue string) {
	_, err := coerceBoolean(rawValue)
	if err != nil {
		result.AddError(name, "type", err.Error())
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
