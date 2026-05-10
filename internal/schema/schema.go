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
	Type        Type    `yaml:"type"`
	Required    bool    `yaml:"required,omitempty"`
	Default     any     `yaml:"default,omitempty"`
	Description string  `yaml:"description,omitempty"`
	Pattern     string  `yaml:"pattern,omitempty"`
	Enum        []any   `yaml:"enum,omitempty"`
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
