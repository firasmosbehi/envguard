package dotenv

import (
	"strings"
	"testing"
)

func TestExpandValueEscapedBrace(t *testing.T) {
	vars := map[string]string{
		"NAME": "world",
		"LIT":  `\{NAME}`,
	}

	if err := Expand(vars); err != nil {
		t.Fatalf("Expand failed: %v", err)
	}

	if vars["LIT"] != `\{NAME}` {
		t.Errorf("expected '\\{NAME}', got %q", vars["LIT"])
	}
}

func TestExpandValueMissingClosingBrace(t *testing.T) {
	vars := map[string]string{
		"REF": "${UNclosed",
	}

	if err := Expand(vars); err != nil {
		t.Fatalf("Expand failed: %v", err)
	}

	if vars["REF"] != "${UNclosed" {
		t.Errorf("expected '${UNclosed}', got %q", vars["REF"])
	}
}

func TestExpandValueNestedExpansion(t *testing.T) {
	vars := map[string]string{
		"BASE":    "hello",
		"WRAPPED": "${BASE}",
		"FINAL":   "${WRAPPED} world",
	}

	if err := Expand(vars); err != nil {
		t.Fatalf("Expand failed: %v", err)
	}

	if vars["FINAL"] != "hello world" {
		t.Errorf("expected 'hello world', got %q", vars["FINAL"])
	}
}

func TestFindClosingBraceNested(t *testing.T) {
	// nested braces
	s := "outer ${INNER} rest"
	idx := findClosingBrace(s, 8)
	if idx != 13 {
		t.Errorf("expected index 13, got %d", idx)
	}

	// no closing brace
	s2 := "no closing ${ brace"
	idx2 := findClosingBrace(s2, 12)
	if idx2 != -1 {
		t.Errorf("expected -1, got %d", idx2)
	}
}

func TestExpandExprDefaultNested(t *testing.T) {
	vars := map[string]string{
		"A": "value",
		"B": "${A:-fallback}",
	}

	if err := Expand(vars); err != nil {
		t.Fatalf("Expand failed: %v", err)
	}

	if vars["B"] != "value" {
		t.Errorf("expected 'value', got %q", vars["B"])
	}
}

func TestExpandExprDefaultWithNestedRef(t *testing.T) {
	vars := map[string]string{
		"A": "inner",
		"B": "${UNDEFINED:-${A}}",
	}

	if err := Expand(vars); err != nil {
		t.Fatalf("Expand failed: %v", err)
	}

	if vars["B"] != "inner" {
		t.Errorf("expected 'inner', got %q", vars["B"])
	}
}

func TestExpandExprErrorEmptyMessage(t *testing.T) {
	vars := map[string]string{
		"REF": "${UNDEFINED:?}",
	}

	err := Expand(vars)
	if err == nil {
		t.Fatal("expected error for required variable")
	}
	if !strings.Contains(err.Error(), `variable "REF"`) {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestUnescapeDoubleQuotes(t *testing.T) {
	// Test various escape sequences
	tests := []struct {
		input string
		want  string
	}{
		{`hello \"world\"`, `hello "world"`},
		{`line1\nline2`, "line1\nline2"},
		{`tab\there`, "tab\there"},
		{`back\\slash`, `back\slash`},
		{`carriage\rreturn`, "carriage\rreturn"},
		{`normal`, `normal`},
	}

	for _, tt := range tests {
		got := unescapeDoubleQuotes(tt.input)
		if got != tt.want {
			t.Errorf("unescapeDoubleQuotes(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
