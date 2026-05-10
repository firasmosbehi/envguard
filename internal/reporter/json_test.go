package reporter

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/envguard/envguard/internal/validator"
)

func TestJSONValid(t *testing.T) {
	result := validator.NewResult()
	var buf bytes.Buffer
	if err := JSON(&buf, result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded validator.Result
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if !decoded.Valid {
		t.Error("expected Valid = true")
	}
	if len(decoded.Errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(decoded.Errors))
	}
}

func TestJSONWithErrors(t *testing.T) {
	result := validator.NewResult()
	result.AddError("FOO", "required", "missing")
	result.AddError("BAR", "type", "invalid")

	var buf bytes.Buffer
	if err := JSON(&buf, result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded validator.Result
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if decoded.Valid {
		t.Error("expected Valid = false")
	}
	if len(decoded.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(decoded.Errors))
	}
}
