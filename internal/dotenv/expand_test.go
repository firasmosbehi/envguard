package dotenv

import (
	"strings"
	"testing"
)

func TestExpandBasic(t *testing.T) {
	vars := map[string]string{
		"NAME":     "world",
		"GREETING": "hello ${NAME}",
	}

	if err := Expand(vars); err != nil {
		t.Fatalf("Expand failed: %v", err)
	}

	if vars["GREETING"] != "hello world" {
		t.Errorf("expected 'hello world', got %q", vars["GREETING"])
	}
}

func TestExpandDefault(t *testing.T) {
	vars := map[string]string{
		"HOST":    "localhost",
		"PORT":    "",
		"URL":     "http://${HOST}:${PORT:-3000}",
		"MISSING": "${UNDEFINED:-fallback}",
	}

	if err := Expand(vars); err != nil {
		t.Fatalf("Expand failed: %v", err)
	}

	if vars["URL"] != "http://localhost:3000" {
		t.Errorf("expected 'http://localhost:3000', got %q", vars["URL"])
	}
	if vars["MISSING"] != "fallback" {
		t.Errorf("expected 'fallback', got %q", vars["MISSING"])
	}
}

func TestExpandRequired(t *testing.T) {
	vars := map[string]string{
		"REQUIRED": "${MUST_BE_SET:?is required}",
	}

	err := Expand(vars)
	if err == nil {
		t.Error("expected error for unset required variable")
	}
	if err != nil && err.Error() != `variable "REQUIRED": is required` {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestExpandEscape(t *testing.T) {
	vars := map[string]string{
		"NAME":    "world",
		"LITERAL": `\${NAME}`,
		"MIXED":   `\${NAME} and ${NAME}`,
	}

	if err := Expand(vars); err != nil {
		t.Fatalf("Expand failed: %v", err)
	}

	if vars["LITERAL"] != "${NAME}" {
		t.Errorf("expected '${NAME}', got %q", vars["LITERAL"])
	}
	if vars["MIXED"] != "${NAME} and world" {
		t.Errorf("expected '${NAME} and world', got %q", vars["MIXED"])
	}
}

func TestExpandNested(t *testing.T) {
	vars := map[string]string{
		"A": "a-value",
		"B": "${A}",
		"C": "${B}",
	}

	if err := Expand(vars); err != nil {
		t.Fatalf("Expand failed: %v", err)
	}

	if vars["C"] != "a-value" {
		t.Errorf("expected 'a-value', got %q", vars["C"])
	}
}

func TestExpandCircular(t *testing.T) {
	vars := map[string]string{
		"A": "${B}",
		"B": "${A}",
	}

	err := Expand(vars)
	if err == nil {
		t.Error("expected error for circular reference")
	}
	if err != nil && !strings.Contains(err.Error(), "circular reference detected") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestExpandSelfReference(t *testing.T) {
	vars := map[string]string{
		"PATH": "${PATH}:/extra",
	}

	err := Expand(vars)
	if err == nil {
		t.Error("expected error for self-reference")
	}
}

func TestExpandNotSet(t *testing.T) {
	vars := map[string]string{
		"REF": "${UNDEFINED}",
	}

	if err := Expand(vars); err != nil {
		t.Fatalf("Expand failed: %v", err)
	}

	if vars["REF"] != "" {
		t.Errorf("expected empty string, got %q", vars["REF"])
	}
}
