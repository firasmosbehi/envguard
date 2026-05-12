package reporter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/envguard/envguard/internal/validator"
)

func TestGitHubValid(t *testing.T) {
	result := validator.NewResult()
	var buf bytes.Buffer
	GitHub(&buf, result, []string{".env"})
	if !strings.Contains(buf.String(), "✓ All environment variables validated.") {
		t.Errorf("expected success message, got: %s", buf.String())
	}
}

func TestGitHubErrors(t *testing.T) {
	result := validator.NewResult()
	result.AddError("FOO", "required", "variable is missing")

	var buf bytes.Buffer
	GitHub(&buf, result, []string{".env"})
	out := buf.String()

	if !strings.Contains(out, "::error title=EnvGuard Validation Error::") {
		t.Errorf("expected GitHub error annotation, got: %s", out)
	}
	if !strings.Contains(out, "FOO=required: variable is missing") {
		t.Errorf("expected FOO error, got: %s", out)
	}
}

func TestGitHubWarnings(t *testing.T) {
	result := validator.NewResult()
	result.AddWarning("UNKNOWN", "strict", "not defined in schema")

	var buf bytes.Buffer
	GitHub(&buf, result, []string{".env"})
	out := buf.String()

	if !strings.Contains(out, "::warning title=EnvGuard Validation Warning::") {
		t.Errorf("expected GitHub warning annotation, got: %s", out)
	}
}
