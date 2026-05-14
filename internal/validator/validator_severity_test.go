package validator

import (
	"testing"

	"github.com/envguard/envguard/internal/schema"
)

func TestSeverityLevels(t *testing.T) {
	tests := []struct {
		name           string
		severity       string
		wantValid      bool
		wantFailOnWarn bool
	}{
		{"error severity fails validation", "error", false, false},
		{"warn severity passes validation", "warn", true, false},
		{"info severity passes validation", "info", true, false},
		{"empty severity defaults to error", "", false, false},
		{"warn severity fails with failOnWarnings", "warn", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &schema.Schema{
				Version: "1.0",
				Env: map[string]*schema.Variable{
					"PORT": {
						Type:     schema.TypeInteger,
						Required: true,
						Severity: tt.severity,
					},
				},
			}

			// PORT is missing, so validation should fail according to severity
			envVars := map[string]string{}
			result := Validate(s, envVars, false, "")

			if result.Valid != tt.wantValid {
				t.Errorf("expected Valid=%v, got %v", tt.wantValid, result.Valid)
			}

			if result.IsValid(tt.wantFailOnWarn) == result.Valid && tt.wantFailOnWarn {
				t.Errorf("expected IsValid(%v)=false, got %v", tt.wantFailOnWarn, result.IsValid(tt.wantFailOnWarn))
			}

			// Check that the error has the correct severity
			if len(result.Errors) != 1 {
				t.Fatalf("expected 1 error, got %d", len(result.Errors))
			}

			var expectedSeverity Severity
			switch tt.severity {
			case "warn":
				expectedSeverity = SeverityWarn
			case "info":
				expectedSeverity = SeverityInfo
			default:
				expectedSeverity = SeverityError
			}

			if result.Errors[0].Severity != expectedSeverity {
				t.Errorf("expected severity %v, got %v", expectedSeverity, result.Errors[0].Severity)
			}
		})
	}
}

func TestSeverityWarnWithValidValue(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"PORT": {
				Type:     schema.TypeInteger,
				Required: true,
				Severity: "warn",
			},
		},
	}

	envVars := map[string]string{"PORT": "8080"}
	result := Validate(s, envVars, false, "")

	if !result.Valid {
		t.Errorf("expected valid result, got errors: %v", result.Errors)
	}
}

func TestSeverityMixedErrorsAndWarnings(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"DATABASE_URL": {
				Type:     schema.TypeString,
				Required: true,
				Severity: "error",
			},
			"OPTIONAL_VAR": {
				Type:     schema.TypeString,
				Required: true,
				Severity: "warn",
			},
		},
	}

	envVars := map[string]string{}
	result := Validate(s, envVars, false, "")

	if result.Valid {
		t.Error("expected invalid result due to error severity")
	}

	if len(result.Errors) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(result.Errors))
	}

	errorCount := 0
	warnCount := 0
	for _, e := range result.Errors {
		switch e.Severity {
		case SeverityError:
			errorCount++
		case SeverityWarn:
			warnCount++
		}
	}

	if errorCount != 1 {
		t.Errorf("expected 1 error severity, got %d", errorCount)
	}
	if warnCount != 1 {
		t.Errorf("expected 1 warn severity, got %d", warnCount)
	}

	// With failOnWarnings=true, should be invalid
	if result.IsValid(true) {
		t.Error("expected IsValid(true)=false when there are warnings")
	}
}
