// Package schema provides parsing and validation of EnvGuard YAML schema files.
package schema

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Type represents the data type of an environment variable.
type Type string

// Supported variable data types.
const (
	TypeString  Type = "string"
	TypeInteger Type = "integer"
	TypeFloat   Type = "float"
	TypeBoolean Type = "boolean"
	TypeArray   Type = "array"
)

// ValidTypes is the set of supported types.
var ValidTypes = map[Type]bool{
	TypeString:  true,
	TypeInteger: true,
	TypeFloat:   true,
	TypeBoolean: true,
	TypeArray:   true,
}

// Variable defines the schema for a single environment variable.
type Variable struct {
	Type        Type     `yaml:"type"`
	Required    bool     `yaml:"required,omitempty"`
	Default     any      `yaml:"default,omitempty"`
	Description string   `yaml:"description,omitempty"`
	Message     string   `yaml:"message,omitempty"`
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
	Separator   string   `yaml:"separator,omitempty"`
	AllowEmpty  *bool    `yaml:"allowEmpty,omitempty"`
	Contains    string   `yaml:"contains,omitempty"`
	DependsOn   string   `yaml:"dependsOn,omitempty"`
	When        string   `yaml:"when,omitempty"`
	Deprecated  string   `yaml:"deprecated,omitempty"`
	Sensitive   bool     `yaml:"sensitive,omitempty"`
	Transform   string   `yaml:"transform,omitempty"`
	Severity    string   `yaml:"severity,omitempty"`
	Prefix      string   `yaml:"prefix,omitempty"`
	Suffix      string   `yaml:"suffix,omitempty"`
	ItemType    Type     `yaml:"itemType,omitempty"`
	UniqueItems bool     `yaml:"uniqueItems,omitempty"`
	ItemPattern string   `yaml:"itemPattern,omitempty"`
	NotEmpty    *bool    `yaml:"notEmpty,omitempty"`
	MultipleOf  any      `yaml:"multipleOf,omitempty"`
}

// CustomSecretRule defines a user-provided secret detection pattern.
type CustomSecretRule struct {
	Name     string `yaml:"name"`
	Pattern  string `yaml:"pattern"`
	Message  string `yaml:"message"`
	Severity string `yaml:"severity,omitempty"`
}

// Secrets defines optional custom secret detection rules.
type Secrets struct {
	Custom []CustomSecretRule `yaml:"custom,omitempty"`
}

// Schema is the top-level structure of an envguard.yaml file.
type Schema struct {
	Version string               `yaml:"version"`
	Extends string               `yaml:"extends,omitempty"`
	Env     map[string]*Variable `yaml:"env"`
	Secrets *Secrets             `yaml:"secrets,omitempty"`
}

// Parse reads and parses a schema YAML file from the given path.
// If the schema has an `extends` field, the base schema is loaded and merged first.
func Parse(path string) (*Schema, error) {
	if cached, ok := DefaultSchemaCache.Get(path); ok {
		return cached, nil
	}
	s, err := parseWithStack(path, make(map[string]bool))
	if err != nil {
		return nil, err
	}
	DefaultSchemaCache.Set(path, s)
	return s, nil
}

// ParseLenient reads and parses a schema YAML file without validating it.
// Useful for tools that need to inspect potentially invalid schemas.
func ParseLenient(path string) (*Schema, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve schema path %s: %w", path, err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file %s: %w", path, err)
	}

	var s Schema
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("failed to parse schema file %s: %w", path, err)
	}

	// Load and merge base schema if extends is set
	if s.Extends != "" {
		basePath := s.Extends
		if !filepath.IsAbs(basePath) {
			basePath = filepath.Join(filepath.Dir(absPath), basePath)
		}
		base, err := ParseLenient(basePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load base schema %q: %w", s.Extends, err)
		}
		s = *mergeSchemas(base, &s)
	}

	return &s, nil
}

func parseWithStack(path string, stack map[string]bool) (*Schema, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve schema path %s: %w", path, err)
	}

	if stack[absPath] {
		return nil, fmt.Errorf("circular schema inheritance detected: %s", path)
	}
	stack[absPath] = true

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file %s: %w", path, err)
	}

	var s Schema
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("failed to parse schema file %s: %w", path, err)
	}

	// Load and merge base schema if extends is set
	if s.Extends != "" {
		basePath := s.Extends
		if !filepath.IsAbs(basePath) {
			basePath = filepath.Join(filepath.Dir(absPath), basePath)
		}
		base, err := parseWithStack(basePath, stack)
		if err != nil {
			return nil, fmt.Errorf("failed to load base schema %q: %w", s.Extends, err)
		}
		s = *mergeSchemas(base, &s)
	}

	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("invalid schema: %w", err)
	}

	return &s, nil
}

// mergeSchemas merges child into base. Child values override base values.
func mergeSchemas(base, child *Schema) *Schema {
	merged := &Schema{
		Version: child.Version,
		Env:     make(map[string]*Variable),
	}
	for name, v := range base.Env {
		merged.Env[name] = v
	}
	for name, v := range child.Env {
		merged.Env[name] = v
	}
	return merged
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

	if s.Secrets != nil {
		for i, rule := range s.Secrets.Custom {
			if rule.Name == "" {
				return fmt.Errorf("secret rule %d: name is required", i)
			}
			if rule.Pattern == "" {
				return fmt.Errorf("secret rule %q: pattern is required", rule.Name)
			}
			if _, err := regexp.Compile(rule.Pattern); err != nil {
				return fmt.Errorf("secret rule %q has invalid pattern: %w", rule.Name, err)
			}
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

	if v.MinLength != nil && v.Type != TypeString && v.Type != TypeArray {
		return fmt.Errorf("variable %q: minLength can only be used with string or array type", name)
	}

	if v.MaxLength != nil && v.Type != TypeString && v.Type != TypeArray {
		return fmt.Errorf("variable %q: maxLength can only be used with string or array type", name)
	}

	if v.MinLength != nil && v.MaxLength != nil && *v.MinLength > *v.MaxLength {
		return fmt.Errorf("variable %q: minLength (%d) cannot be greater than maxLength (%d)", name, *v.MinLength, *v.MaxLength)
	}

	if v.Format != "" && v.Type != TypeString {
		return fmt.Errorf("variable %q: format can only be used with string type", name)
	}

	validFormats := map[string]bool{"email": true, "url": true, "uuid": true, "base64": true, "ip": true, "port": true, "json": true, "duration": true, "semver": true, "hostname": true, "hex": true, "cron": true, "datetime": true, "date": true, "time": true, "timezone": true, "color": true, "slug": true, "filepath": true, "directory": true, "locale": true, "jwt": true, "mongodb-uri": true, "redis-uri": true}
	if v.Format != "" && !validFormats[v.Format] {
		return fmt.Errorf("variable %q: unsupported format %q (supported: email, url, uuid, base64, ip, port, json, duration, semver, hostname, hex, cron, datetime, date, time, timezone, color, slug, filepath, directory, locale, jwt, mongodb-uri, redis-uri)", name, v.Format)
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

	if v.Separator != "" && v.Type != TypeArray {
		return fmt.Errorf("variable %q: separator can only be used with array type", name)
	}

	if v.Type == TypeArray && v.Separator == "" {
		return fmt.Errorf("variable %q: array type requires a separator", name)
	}

	if v.Contains != "" && v.Type != TypeArray {
		return fmt.Errorf("variable %q: contains can only be used with array type", name)
	}

	if v.DependsOn != "" && v.When == "" {
		return fmt.Errorf("variable %q: dependsOn requires when", name)
	}

	if v.When != "" && v.DependsOn == "" {
		return fmt.Errorf("variable %q: when requires dependsOn", name)
	}

	if v.AllowEmpty != nil && !*v.AllowEmpty && v.Required {
		return fmt.Errorf("variable %q: allowEmpty=false is redundant when required=true", name)
	}

	if v.Transform != "" && v.Type != TypeString {
		return fmt.Errorf("variable %q: transform can only be used with string type", name)
	}

	validTransforms := map[string]bool{"lowercase": true, "uppercase": true, "trim": true}
	if v.Transform != "" && !validTransforms[v.Transform] {
		return fmt.Errorf("variable %q: unsupported transform %q (supported: lowercase, uppercase, trim)", name, v.Transform)
	}

	validSeverities := map[string]bool{"error": true, "warn": true, "info": true}
	if v.Severity != "" && !validSeverities[v.Severity] {
		return fmt.Errorf("variable %q: unsupported severity %q (supported: error, warn, info)", name, v.Severity)
	}

	if v.Prefix != "" && v.Type != TypeString {
		return fmt.Errorf("variable %q: prefix can only be used with string type", name)
	}

	if v.Suffix != "" && v.Type != TypeString {
		return fmt.Errorf("variable %q: suffix can only be used with string type", name)
	}

	if v.ItemType != "" && v.Type != TypeArray {
		return fmt.Errorf("variable %q: itemType can only be used with array type", name)
	}
	if v.ItemType != "" && !ValidTypes[v.ItemType] {
		return fmt.Errorf("variable %q: unsupported itemType %q", name, v.ItemType)
	}

	if v.UniqueItems && v.Type != TypeArray {
		return fmt.Errorf("variable %q: uniqueItems can only be used with array type", name)
	}

	if v.ItemPattern != "" && v.Type != TypeArray {
		return fmt.Errorf("variable %q: itemPattern can only be used with array type", name)
	}
	if v.ItemPattern != "" {
		if _, err := regexp.Compile(v.ItemPattern); err != nil {
			return fmt.Errorf("variable %q has invalid itemPattern: %w", name, err)
		}
	}

	if v.NotEmpty != nil && v.Type != TypeArray {
		return fmt.Errorf("variable %q: notEmpty can only be used with array type", name)
	}

	if v.MultipleOf != nil {
		if v.Type != TypeInteger && v.Type != TypeFloat {
			return fmt.Errorf("variable %q: multipleOf can only be used with integer or float type", name)
		}
		if err := validateNumeric(name, "multipleOf", v.MultipleOf); err != nil {
			return err
		}
		if val, ok := toFloat64(v.MultipleOf); !ok || val == 0 {
			return fmt.Errorf("variable %q: multipleOf must be a non-zero number", name)
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
		switch v := value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return nil
		case float64:
			// yaml.v3 unmarshals numbers as float64; check if it's a whole number
			if v == float64(int64(v)) {
				return nil
			}
			return fmt.Errorf("enum value must be an integer, got %v", value)
		default:
			return fmt.Errorf("enum value must be an integer, got %T", value)
		}
	case TypeFloat:
		switch v := value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			_ = v
			return nil
		default:
			return fmt.Errorf("enum value must be a number, got %T", value)
		}
	case TypeArray:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("enum value must be a string, got %T", value)
		}
		return nil
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
		switch v := value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return nil
		case float64:
			if v == float64(int64(v)) {
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
		//nolint:staticcheck // Original form is more readable than De Morgan's equivalent.
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
