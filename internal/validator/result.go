// Package validator provides the validation engine for EnvGuard.
package validator

// ValidationError represents a single validation failure.
type ValidationError struct {
	Key     string `json:"key"`
	Message string `json:"message"`
	Rule    string `json:"rule"`
}

// Result holds the outcome of a validation run.
type Result struct {
	Valid    bool               `json:"valid"`
	Errors   []ValidationError  `json:"errors"`
	Warnings []ValidationError  `json:"warnings"`
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
