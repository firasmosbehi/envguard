// Package envguard provides a public Go API for validating .env files against schemas.
package envguard

import (
	"github.com/envguard/envguard/internal/dotenv"
	"github.com/envguard/envguard/internal/schema"
	"github.com/envguard/envguard/internal/validator"
)

// ValidationError represents a single validation failure.
type ValidationError = validator.ValidationError

// Result holds the outcome of a validation run.
type Result = validator.Result

// Schema represents a parsed EnvGuard schema.
type Schema = schema.Schema

// Variable represents a single environment variable definition.
type Variable = schema.Variable

// Type represents the data type of an environment variable.
type Type = schema.Type

// Supported type constants.
const (
	TypeString  = schema.TypeString
	TypeInteger = schema.TypeInteger
	TypeFloat   = schema.TypeFloat
	TypeBoolean = schema.TypeBoolean
	TypeArray   = schema.TypeArray
)

// ParseSchema reads and parses a schema YAML file.
func ParseSchema(path string) (*Schema, error) {
	return schema.Parse(path)
}

// ParseEnv reads a .env file into a map of variable names to values.
func ParseEnv(path string) (map[string]string, error) {
	return dotenv.Parse(path)
}

// Validate checks envVars against the given schema.
// If strict is true, warnings are generated for keys in envVars not defined in the schema.
// envName is the current environment (e.g. "production") for environment-specific rules.
func Validate(s *Schema, envVars map[string]string, strict bool, envName string) *Result {
	return validator.Validate(s, envVars, strict, envName)
}

// ValidateFile validates a .env file against a schema file.
// This is a convenience wrapper around ParseSchema, ParseEnv, and Validate.
func ValidateFile(schemaPath, envPath string, strict bool, envName string) (*Result, error) {
	s, err := ParseSchema(schemaPath)
	if err != nil {
		return nil, err
	}

	envVars, err := ParseEnv(envPath)
	if err != nil {
		return nil, err
	}

	return Validate(s, envVars, strict, envName), nil
}
