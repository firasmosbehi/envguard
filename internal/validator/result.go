// Package validator provides the validation engine for EnvGuard.
package validator

import (
	"strings"

	"github.com/envguard/envguard/internal/schema"
)

// Severity represents the severity level of a validation issue.
type Severity string

const (
	SeverityError Severity = "error"
	SeverityWarn  Severity = "warn"
	SeverityInfo  Severity = "info"
)

// ValidationError represents a single validation failure.
type ValidationError struct {
	Key      string   `json:"key"`
	Message  string   `json:"message"`
	Rule     string   `json:"rule"`
	Severity Severity `json:"severity,omitempty"`
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

// AddError records a validation error with severity and marks the result as invalid if severity is error.
func (r *Result) AddError(key, rule, message string) {
	r.AddErrorWithSeverity(key, rule, message, SeverityError)
}

// AddErrorWithSeverity records a validation error with the given severity.
// Only severity "error" marks the result as invalid.
func (r *Result) AddErrorWithSeverity(key, rule, message string, severity Severity) {
	if severity == SeverityError {
		r.Valid = false
	}
	r.Errors = append(r.Errors, ValidationError{
		Key:      key,
		Rule:     rule,
		Message:  message,
		Severity: severity,
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

// IsValid returns whether the result is considered valid.
// If failOnWarnings is true, warnings are treated as errors.
func (r *Result) IsValid(failOnWarnings bool) bool {
	if !r.Valid {
		return false
	}
	if failOnWarnings {
		if len(r.Warnings) > 0 {
			return false
		}
		for _, e := range r.Errors {
			if e.Severity == SeverityWarn {
				return false
			}
		}
	}
	return true
}

// HasErrors returns true if there are errors with the given severity or higher.
func (r *Result) HasErrors(minSeverity Severity) bool {
	for _, e := range r.Errors {
		if e.Severity == minSeverity || (minSeverity == SeverityWarn && e.Severity == SeverityError) {
			return true
		}
	}
	return false
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
