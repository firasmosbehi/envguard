// Package validator provides the validation engine for EnvGuard.
package validator

import (
	"strings"

	"github.com/envguard/envguard/internal/schema"
)

// ValidationError represents a single validation failure.
type ValidationError struct {
	Key     string `json:"key"`
	Message string `json:"message"`
	Rule    string `json:"rule"`
}

// Result holds the outcome of a validation run.
type Result struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationError `json:"errors"`
	Warnings []ValidationError `json:"warnings"`
}

// NewResult creates an empty Result.
func NewResult() *Result {
	return &Result{
		Valid:    true,
		Errors:   make([]ValidationError, 0),
		Warnings: make([]ValidationError, 0),
	}
}

// AddError records a validation error and marks the result as invalid.
func (r *Result) AddError(key, rule, message string) {
	r.Valid = false
	r.Errors = append(r.Errors, ValidationError{
		Key:     key,
		Rule:    rule,
		Message: message,
	})
}

// AddWarning records a non-fatal warning.
func (r *Result) AddWarning(key, rule, message string) {
	r.Warnings = append(r.Warnings, ValidationError{
		Key:     key,
		Rule:    rule,
		Message: message,
	})
}

// ErrorCount returns the number of validation errors.
func (r *Result) ErrorCount() int {
	return len(r.Errors)
}

// WarningCount returns the number of warnings.
func (r *Result) WarningCount() int {
	return len(r.Warnings)
}

// RedactSensitive replaces values of sensitive variables in error/warning messages with ***.
func (r *Result) RedactSensitive(envVars map[string]string, s *schema.Schema) {
	for name, variable := range s.Env {
		if !variable.Sensitive {
			continue
		}
		value, exists := envVars[name]
		if !exists || value == "" {
			continue
		}
		for i := range r.Errors {
			r.Errors[i].Message = strings.ReplaceAll(r.Errors[i].Message, value, "***")
		}
		for i := range r.Warnings {
			r.Warnings[i].Message = strings.ReplaceAll(r.Warnings[i].Message, value, "***")
		}
	}
}
