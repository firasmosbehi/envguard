// Package schema provides parsing and validation of EnvGuard YAML schema files.
package schema

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Type represents the data type of an environment variable.
type Type string

const (
	TypeString  Type = "string"
	TypeInteger Type = "integer"
	TypeFloat   Type = "float"
	TypeBoolean Type = "boolean"
)

// ValidTypes is the set of supported types.
var ValidTypes = map[Type]bool{
	TypeString:  true,
	TypeInteger: true,
	TypeFloat:   true,
	TypeBoolean: true,
}

// Variable defines the schema for a single environment variable.
type Variable struct {
	Type        Type     `yaml:"type"`
	Required    bool     `yaml:"required,omitempty"`
	Default     any      `yaml:"default,omitempty"`
	Description string   `yaml:"description,omitempty"`
	Pattern     string   `yaml:"pattern,omitempty"`
	Enum        []any    `yaml:"enum,omitempty"`
	Min         any      `yaml:"min,omitempty"`
	Max         any      `yaml:"max,omitempty"`
	MinLength   *int     `yaml:"minLength,omitempty"`
	MaxLength   *int     `yaml:"maxLength,omitempty"`
	Format      string   `yaml:"format,omitempty"`
	Disallow    []string `yaml:"disallow,omitempty"`
	RequiredIn  []string `yaml:"requiredIn,omitempty"`
	DevOnly     bool     `yaml:"devOnly,omitempty"`
}

// Schema is the top-level structure of an envguard.yaml file.
type Schema struct {
	Version string                `yaml:"version"`
	Env     map[string]*Variable  `yaml:"env"`
}

// Parse reads and parses a schema YAML file from the given path.
func Parse(path string) (*Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file %s: %w", path, err)
	}

	var s Schema
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("failed to parse schema file %s: %w", path, err)
	}

	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("invalid schema: %w", err)
	}

	return &s, nil
}

// Validate checks the schema for structural correctness.
func (s *Schema) Validate() error {
	if s.Version == "" {
		return fmt.Errorf("schema version is required")
	}

	if len(s.Env) == 0 {
		return fmt.Errorf("schema must define at least one environment variable")
	}

	for name, v := range s.Env {
		if err := validateVariable(name, v); err != nil {
			return err
		}
	}

	return nil
}

func validateVariable(name string, v *Variable) error {
	if v == nil {
		return fmt.Errorf("variable %q has no definition", name)
	}

	if !ValidTypes[v.Type] {
		return fmt.Errorf("variable %q has unsupported type %q", name, v.Type)
	}

	if v.Required && v.Default != nil {
		return fmt.Errorf("variable %q: required and default are mutually exclusive", name)
	}

	if v.Pattern != "" && v.Type != TypeString {
		return fmt.Errorf("variable %q: pattern can only be used with string type", name)
	}

	if v.Pattern != "" {
		if _, err := regexp.Compile(v.Pattern); err != nil {
			return fmt.Errorf("variable %q has invalid pattern: %w", name, err)
		}
	}

	if v.Enum != nil {
		if len(v.Enum) == 0 {
			return fmt.Errorf("variable %q: enum cannot be empty", name)
		}
		if v.Type == TypeBoolean {
			return fmt.Errorf("variable %q: enum cannot be used with boolean type", name)
		}
		for _, ev := range v.Enum {
			if err := validateEnumValue(v.Type, ev); err != nil {
				return fmt.Errorf("variable %q has invalid enum value: %w", name, err)
			}
		}
	}

	if v.Min != nil {
		if v.Type != TypeInteger && v.Type != TypeFloat {
			return fmt.Errorf("variable %q: min can only be used with integer or float type", name)
		}
		if err := validateNumeric(name, "min", v.Min); err != nil {
			return err
		}
	}

	if v.Max != nil {
		if v.Type != TypeInteger && v.Type != TypeFloat {
			return fmt.Errorf("variable %q: max can only be used with integer or float type", name)
		}
		if err := validateNumeric(name, "max", v.Max); err != nil {
			return err
		}
	}

	if v.Min != nil && v.Max != nil {
		minVal, _ := toFloat64(v.Min)
		maxVal, _ := toFloat64(v.Max)
		if minVal > maxVal {
			return fmt.Errorf("variable %q: min (%v) cannot be greater than max (%v)", name, v.Min, v.Max)
		}
	}

	if v.MinLength != nil && v.Type != TypeString {
		return fmt.Errorf("variable %q: minLength can only be used with string type", name)
	}

	if v.MaxLength != nil && v.Type != TypeString {
		return fmt.Errorf("variable %q: maxLength can only be used with string type", name)
	}

	if v.MinLength != nil && v.MaxLength != nil && *v.MinLength > *v.MaxLength {
		return fmt.Errorf("variable %q: minLength (%d) cannot be greater than maxLength (%d)", name, *v.MinLength, *v.MaxLength)
	}

	if v.Format != "" && v.Type != TypeString {
		return fmt.Errorf("variable %q: format can only be used with string type", name)
	}

	validFormats := map[string]bool{"email": true, "url": true, "uuid": true}
	if v.Format != "" && !validFormats[v.Format] {
		return fmt.Errorf("variable %q: unsupported format %q (supported: email, url, uuid)", name, v.Format)
	}

	if len(v.Disallow) > 0 && v.Type != TypeString {
		return fmt.Errorf("variable %q: disallow can only be used with string type", name)
	}

	if v.DevOnly && v.Required {
		return fmt.Errorf("variable %q: devOnly and required are mutually exclusive", name)
	}

	if v.DevOnly && len(v.RequiredIn) > 0 {
		return fmt.Errorf("variable %q: devOnly and requiredIn are mutually exclusive", name)
	}

	if v.Default != nil {
		if err := validateDefault(v.Type, v.Default); err != nil {
			return fmt.Errorf("variable %q has invalid default: %w", name, err)
		}
	}

	return nil
}

func validateEnumValue(t Type, value any) error {
	switch t {
	case TypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("enum value must be a string, got %T", value)
		}
	case TypeInteger:
		// YAML may parse integers as int, but unmarshal into interface{} gives various numeric types
		switch value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return nil
		case float64:
			// yaml.v3 unmarshals numbers as float64; check if it's a whole number
			if v, ok := value.(float64); ok && v == float64(int64(v)) {
				return nil
			}
			return fmt.Errorf("enum value must be an integer, got %v", value)
		default:
			return fmt.Errorf("enum value must be an integer, got %T", value)
		}
	case TypeFloat:
		switch value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			return nil
		default:
			return fmt.Errorf("enum value must be a number, got %T", value)
		}
	default:
		return fmt.Errorf("enum not supported for type %s", t)
	}
	return nil
}

func validateDefault(t Type, value any) error {
	switch t {
	case TypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("default must be a string, got %T", value)
		}
	case TypeInteger:
		switch value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return nil
		case float64:
			if v, ok := value.(float64); ok && v == float64(int64(v)) {
				return nil
			}
			return fmt.Errorf("default must be an integer, got %v", value)
		default:
			return fmt.Errorf("default must be an integer, got %T", value)
		}
	case TypeFloat:
		switch value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			return nil
		default:
			return fmt.Errorf("default must be a number, got %T", value)
		}
	case TypeBoolean:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("default must be a boolean, got %T", value)
		}
	}
	return nil
}

func validateNumeric(name, field string, value any) error {
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return nil
	default:
		return fmt.Errorf("variable %q: %s must be a number, got %T", name, field, value)
	}
}

func toFloat64(value any) (float64, bool) {
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

// IsEnvVarNameValid checks if a string is a valid environment variable name.
func IsEnvVarNameValid(name string) bool {
	if name == "" {
		return false
	}
	for i, c := range name {
		if i == 0 && (c >= '0' && c <= '9') {
			return false
		}
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}

// NormalizeName returns the canonical form of an environment variable name.
func NormalizeName(name string) string {
	return strings.TrimSpace(name)
}
