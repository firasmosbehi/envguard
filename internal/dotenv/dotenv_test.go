package dotenv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	content := `
# This is a comment
DATABASE_URL=postgresql://localhost/mydb
PORT=3000

# Empty value
EMPTY=

# Quoted values
QUOTED_DOUBLE="hello world"
QUOTED_SINGLE='hello world'

# Escaped values
ESCAPED="line1\nline2"

# Inline comment handling (not supported, value includes the comment part)
WITH_COMMENT=foo # this is part of value
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	vars, err := Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := map[string]string{
		"DATABASE_URL":  "postgresql://localhost/mydb",
		"PORT":          "3000",
		"EMPTY":         "",
		"QUOTED_DOUBLE": "hello world",
		"QUOTED_SINGLE": "hello world",
		"ESCAPED":       "line1\nline2",
		"WITH_COMMENT":  "foo # this is part of value",
	}

	if len(vars) != len(expected) {
		t.Errorf("expected %d vars, got %d", len(expected), len(vars))
	}

	for key, want := range expected {
		got, ok := vars[key]
		if !ok {
			t.Errorf("missing key %q", key)
			continue
		}
		if got != want {
			t.Errorf("key %q: got %q, want %q", key, got, want)
		}
	}
}

func TestParseMissingFile(t *testing.T) {
	_, err := Parse("/nonexistent/.env")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestParseLine(t *testing.T) {
	tests := []struct {
		line    string
		wantKey string
		wantVal string
		wantErr bool
	}{
		{"FOO=bar", "FOO", "bar", false},
		{"FOO=", "FOO", "", false},
		{"FOO= bar ", "FOO", "bar", false},
		{"FOO=\"hello world\"", "FOO", "hello world", false},
		{"FOO='hello world'", "FOO", "hello world", false},
		{"=bar", "", "", true},
		{"FOO", "", "", false}, // no '=' -> treated as no key
	}

	for _, tt := range tests {
		key, val, err := parseLine(tt.line)
		if tt.wantErr {
			if err == nil {
				t.Errorf("parseLine(%q): expected error", tt.line)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseLine(%q): unexpected error: %v", tt.line, err)
			continue
		}
		if key != tt.wantKey || val != tt.wantVal {
			t.Errorf("parseLine(%q) = (%q, %q), want (%q, %q)",
				tt.line, key, val, tt.wantKey, tt.wantVal)
		}
	}
}

func TestUnquote(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`"hello"`, "hello"},
		{`'hello'`, "hello"},
		{`hello`, "hello"},
		{`""`, ""},
		{`''`, ""},
		{`"line1\nline2"`, "line1\nline2"},
		{`"tab\there"`, "tab\there"},
		{`"backslash\\"`, "backslash\\"},
		{`"quote\"inside"`, "quote\"inside"},
	}

	for _, tt := range tests {
		got := unquote(tt.input)
		if got != tt.want {
			t.Errorf("unquote(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
