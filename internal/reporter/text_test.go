package reporter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/envguard/envguard/internal/validator"
)

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

	if !strings.Contains(out, "⚠ Warnings") {
		t.Errorf("expected warnings header, got: %s", out)
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
	if !strings.Contains(out, "⚠ Warnings") {
		t.Errorf("expected warnings header, got: %s", out)
	}
}
