package validator

import (
	"testing"

	"github.com/envguard/envguard/internal/schema"
)

func TestValidate(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"DATABASE_URL": {
				Type:     schema.TypeString,
				Required: true,
			},
			"PORT": {
				Type:    schema.TypeInteger,
				Default: 3000,
			},
			"DEBUG": {
				Type:    schema.TypeBoolean,
				Default: false,
			},
			"LOG_LEVEL": {
				Type:    schema.TypeString,
				Enum:    []any{"debug", "info", "warn", "error"},
				Default: "info",
			},
			"API_KEY": {
				Type:    schema.TypeString,
				Pattern: "^[A-Za-z0-9_-]{32,}$",
			},
			"RATIO": {
				Type: schema.TypeFloat,
			},
			"COUNT": {
				Type: schema.TypeInteger,
				Enum: []any{1, 2, 3},
			},
			"MODE": {
				Type: schema.TypeFloat,
				Enum: []any{1.5, 2.5, 3.5},
			},
		},
	}

	tests := []struct {
		name      string
		envVars   map[string]string
		strict    bool
		wantValid bool
		wantErrs  int
		wantWarns int
	}{
		{
			name: "valid full env",
			envVars: map[string]string{
				"DATABASE_URL": "postgresql://localhost/mydb",
				"PORT":         "3000",
				"DEBUG":        "false",
				"LOG_LEVEL":    "info",
				"API_KEY":      "abc123def456ghi789jkl012mno345pq",
			},
			wantValid: true,
		},
		{
			name: "valid with defaults",
			envVars: map[string]string{
				"DATABASE_URL": "postgresql://localhost/mydb",
			},
			wantValid: true,
		},
		{
			name: "missing required",
			envVars: map[string]string{
				"PORT": "3000",
			},
			wantValid: false,
			wantErrs:  1,
		},
		{
			name: "empty required",
			envVars: map[string]string{
				"DATABASE_URL": "",
			},
			wantValid: false,
			wantErrs:  1,
		},
		{
			name: "invalid integer",
			envVars: map[string]string{
				"DATABASE_URL": "postgresql://localhost/mydb",
				"PORT":         "not-a-number",
			},
			wantValid: false,
			wantErrs:  1,
		},
		{
			name: "invalid boolean",
			envVars: map[string]string{
				"DATABASE_URL": "postgresql://localhost/mydb",
				"DEBUG":        "maybe",
			},
			wantValid: false,
			wantErrs:  1,
		},
		{
			name: "invalid enum",
			envVars: map[string]string{
				"DATABASE_URL": "postgresql://localhost/mydb",
				"LOG_LEVEL":    "verbose",
			},
			wantValid: false,
			wantErrs:  1,
		},
		{
			name: "invalid pattern",
			envVars: map[string]string{
				"DATABASE_URL": "postgresql://localhost/mydb",
				"API_KEY":      "short",
			},
			wantValid: false,
			wantErrs:  1,
		},
		{
			name: "multiple errors",
			envVars: map[string]string{
				"PORT":      "abc",
				"DEBUG":     "xyz",
				"LOG_LEVEL": "unknown",
			},
			wantValid: false,
			wantErrs:  4, // DATABASE_URL missing + PORT + DEBUG + LOG_LEVEL
		},
		{
			name: "strict mode warns on unknown",
			envVars: map[string]string{
				"DATABASE_URL": "postgresql://localhost/mydb",
				"OLD_VAR":      "value",
			},
			strict:    true,
			wantValid: true,
			wantWarns: 1,
		},
		{
			name: "strict mode no warn when clean",
			envVars: map[string]string{
				"DATABASE_URL": "postgresql://localhost/mydb",
			},
			strict:    true,
			wantValid: true,
			wantWarns: 0,
		},
		{
			name: "valid float",
			envVars: map[string]string{
				"DATABASE_URL": "postgresql://localhost/mydb",
				"RATIO":        "3.14",
			},
			wantValid: true,
		},
		{
			name: "invalid float",
			envVars: map[string]string{
				"DATABASE_URL": "postgresql://localhost/mydb",
				"RATIO":        "abc",
			},
			wantValid: false,
			wantErrs:  1,
		},
		{
			name: "valid integer enum",
			envVars: map[string]string{
				"DATABASE_URL": "postgresql://localhost/mydb",
				"COUNT":        "2",
			},
			wantValid: true,
		},
		{
			name: "invalid integer enum",
			envVars: map[string]string{
				"DATABASE_URL": "postgresql://localhost/mydb",
				"COUNT":        "5",
			},
			wantValid: false,
			wantErrs:  1,
		},
		{
			name: "valid float enum",
			envVars: map[string]string{
				"DATABASE_URL": "postgresql://localhost/mydb",
				"MODE":         "2.5",
			},
			wantValid: true,
		},
		{
			name: "invalid float enum",
			envVars: map[string]string{
				"DATABASE_URL": "postgresql://localhost/mydb",
				"MODE":         "4.5",
			},
			wantValid: false,
			wantErrs:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Validate(s, tt.envVars, tt.strict)
			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", result.Valid, tt.wantValid)
			}
			if len(result.Errors) != tt.wantErrs {
				t.Errorf("Errors count = %d, want %d; errors = %v", len(result.Errors), tt.wantErrs, result.Errors)
			}
			if len(result.Warnings) != tt.wantWarns {
				t.Errorf("Warnings count = %d, want %d; warnings = %v", len(result.Warnings), tt.wantWarns, result.Warnings)
			}
		})
	}
}

func TestResultCounts(t *testing.T) {
	result := NewResult()
	if result.ErrorCount() != 0 {
		t.Errorf("expected 0 errors, got %d", result.ErrorCount())
	}
	if result.WarningCount() != 0 {
		t.Errorf("expected 0 warnings, got %d", result.WarningCount())
	}

	result.AddError("FOO", "required", "missing")
	result.AddWarning("BAR", "strict", "unknown")

	if result.ErrorCount() != 1 {
		t.Errorf("expected 1 error, got %d", result.ErrorCount())
	}
	if result.WarningCount() != 1 {
		t.Errorf("expected 1 warning, got %d", result.WarningCount())
	}
}

func TestDefaultToString(t *testing.T) {
	tests := []struct {
		input any
		want  string
	}{
		{"hello", "hello"},
		{42, "42"},
		{int64(42), "42"},
		{3.14, "3.14"},
		{true, "true"},
		{false, "false"},
	}

	for _, tt := range tests {
		got := defaultToString(tt.input)
		if got != tt.want {
			t.Errorf("defaultToString(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
