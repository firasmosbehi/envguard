package reporter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/envguard/envguard/internal/validator"
)

func TestSeveritySymbol(t *testing.T) {
	tests := []struct {
		sev      validator.Severity
		expected string
	}{
		{validator.SeverityError, "✗"},
		{validator.SeverityWarn, "⚠"},
		{validator.SeverityInfo, "ℹ"},
		{validator.Severity("unknown"), "✗"},
	}

	for _, tt := range tests {
		t.Run(string(tt.sev), func(t *testing.T) {
			got := severitySymbol(tt.sev)
			if got != tt.expected {
				t.Errorf("severitySymbol(%q) = %q, want %q", tt.sev, got, tt.expected)
			}
		})
	}
}

func TestTextValid(t *testing.T) {
	result := validator.NewResult()
	var buf bytes.Buffer
	Text(&buf, result)
	if !strings.Contains(buf.String(), "✓ All environment variables validated.") {
		t.Errorf("expected success message, got: %s", buf.String())
	}
}

func TestTextErrors(t *testing.T) {
	result := validator.NewResult()
	result.AddError("FOO", "required", "variable is missing")
	result.AddError("BAR", "type", "expected integer")

	var buf bytes.Buffer
	Text(&buf, result)
	out := buf.String()

	if !strings.Contains(out, "✗ Environment validation failed") {
		t.Errorf("expected failure header, got: %s", out)
	}
	if !strings.Contains(out, "FOO") {
		t.Errorf("expected FOO error, got: %s", out)
	}
	if !strings.Contains(out, "BAR") {
		t.Errorf("expected BAR error, got: %s", out)
	}
}

func TestTextWarnings(t *testing.T) {
	result := validator.NewResult()
	result.AddWarning("UNKNOWN", "strict", "not defined in schema")

	var buf bytes.Buffer
	Text(&buf, result)
	out := buf.String()

	if !strings.Contains(out, "⚠ General Warnings") {
		t.Errorf("expected general warnings header, got: %s", out)
	}
	if !strings.Contains(out, "UNKNOWN") {
		t.Errorf("expected UNKNOWN warning, got: %s", out)
	}
}

func TestTextErrorsAndWarnings(t *testing.T) {
	result := validator.NewResult()
	result.AddError("FOO", "required", "missing")
	result.AddWarning("UNKNOWN", "strict", "not defined")

	var buf bytes.Buffer
	Text(&buf, result)
	out := buf.String()

	if !strings.Contains(out, "✗ Environment validation failed") {
		t.Errorf("expected failure header, got: %s", out)
	}
	if !strings.Contains(out, "⚠ General Warnings") {
		t.Errorf("expected general warnings header, got: %s", out)
	}
}

func TestText_WarnOnlyErrors(t *testing.T) {
	result := validator.NewResult()
	result.AddErrorWithSeverity("OLD_VAR", "deprecated", "OLD_VAR is deprecated", validator.SeverityWarn)

	var buf bytes.Buffer
	Text(&buf, result)
	out := buf.String()

	if strings.Contains(out, "✗ Environment validation failed") {
		t.Errorf("did not expect failure header for warn-only, got: %s", out)
	}
	if !strings.Contains(out, "⚠ Warnings") {
		t.Errorf("expected warnings header, got: %s", out)
	}
	if !strings.Contains(out, "OLD_VAR") {
		t.Errorf("expected OLD_VAR, got: %s", out)
	}
}

func TestText_InfoOnlyErrors(t *testing.T) {
	result := validator.NewResult()
	result.AddErrorWithSeverity("NOTE_VAR", "info", "this is informational", validator.SeverityInfo)

	var buf bytes.Buffer
	Text(&buf, result)
	out := buf.String()

	if strings.Contains(out, "✗ Environment validation failed") {
		t.Errorf("did not expect failure header for info-only, got: %s", out)
	}
	if !strings.Contains(out, "ℹ Info") {
		t.Errorf("expected info header, got: %s", out)
	}
	if !strings.Contains(out, "NOTE_VAR") {
		t.Errorf("expected NOTE_VAR, got: %s", out)
	}
}

func TestText_MixedSeveritiesAndGeneralWarnings(t *testing.T) {
	result := validator.NewResult()
	result.AddError("ERR", "required", "missing")
	result.AddErrorWithSeverity("WARN", "deprecated", "deprecated", validator.SeverityWarn)
	result.AddErrorWithSeverity("INFO", "suggestion", "suggestion", validator.SeverityInfo)
	result.AddWarning("GENERAL", "strict", "not defined")

	var buf bytes.Buffer
	Text(&buf, result)
	out := buf.String()

	if !strings.Contains(out, "✗ Environment validation failed") {
		t.Errorf("expected failure header, got: %s", out)
	}
	if !strings.Contains(out, "⚠ Warnings") {
		t.Errorf("expected warnings header, got: %s", out)
	}
	if !strings.Contains(out, "ℹ Info") {
		t.Errorf("expected info header, got: %s", out)
	}
	if !strings.Contains(out, "⚠ General Warnings") {
		t.Errorf("expected general warnings header, got: %s", out)
	}
}

func TestText_InvalidWithNoSeverityErrors(t *testing.T) {
	result := validator.NewResult()
	result.AddErrorWithSeverity("WARN", "deprecated", "deprecated", validator.SeverityWarn)
	result.Valid = false // force invalid without severity-error errors

	var buf bytes.Buffer
	Text(&buf, result)
	out := buf.String()

	if !strings.Contains(out, "✗ Environment validation failed") {
		t.Errorf("expected failure header when Valid=false, got: %s", out)
	}
	if !strings.Contains(out, "⚠ Warnings") {
		t.Errorf("expected warnings header, got: %s", out)
	}
}
