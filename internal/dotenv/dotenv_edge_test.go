package dotenv

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// === Edge cases that may expose bugs ===

func TestParseValueWithMultipleEquals(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env")
	os.WriteFile(path, []byte("FOO=bar=baz=qux\n"), 0644)

	vars, err := Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vars["FOO"] != "bar=baz=qux" {
		t.Errorf("expected 'bar=baz=qux', got %q", vars["FOO"])
	}
}

func TestParseVariableWithoutEquals(t *testing.T) {
	// Some parsers treat FOO (no =) as FOO=""
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env")
	os.WriteFile(path, []byte("FOO\n"), 0644)

	vars, err := Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Current behavior: lines without = are skipped
	// If this is a bug, change the implementation
	if _, exists := vars["FOO"]; exists {
		t.Logf("FOO exists with value %q (may or may not be desired)", vars["FOO"])
	}
}

func TestParseQuotedValueWithHash(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env")
	os.WriteFile(path, []byte("FOO=\"value # with hash\"\n"), 0644)

	vars, err := Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vars["FOO"] != "value # with hash" {
		t.Errorf("expected 'value # with hash', got %q", vars["FOO"])
	}
}

func TestParseCRLFLineEndings(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env")
	os.WriteFile(path, []byte("FOO=bar\r\nBAR=baz\r\n"), 0644)

	vars, err := Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vars["FOO"] != "bar" {
		t.Errorf("expected 'bar', got %q", vars["FOO"])
	}
	if vars["BAR"] != "baz" {
		t.Errorf("expected 'baz', got %q", vars["BAR"])
	}
}

func TestParseEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env")
	os.WriteFile(path, []byte(""), 0644)

	vars, err := Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vars) != 0 {
		t.Errorf("expected empty map, got %d entries", len(vars))
	}
}

func TestParseOnlyComments(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env")
	os.WriteFile(path, []byte("# comment 1\n# comment 2\n"), 0644)

	vars, err := Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vars) != 0 {
		t.Errorf("expected empty map, got %d entries", len(vars))
	}
}

func TestParseUnicodeValues(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env")
	os.WriteFile(path, []byte("GREETING=こんにちは\nEMOJI=🚀\n"), 0644)

	vars, err := Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vars["GREETING"] != "こんにちは" {
		t.Errorf("expected 'こんにちは', got %q", vars["GREETING"])
	}
	if vars["EMOJI"] != "🚀" {
		t.Errorf("expected '🚀', got %q", vars["EMOJI"])
	}
}

func TestParseValueWithLeadingTrailingSpaces(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env")
	os.WriteFile(path, []byte("FOO=  bar  \n"), 0644)

	vars, err := Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vars["FOO"] != "bar" {
		t.Errorf("expected 'bar' (trimmed), got %q", vars["FOO"])
	}
}

func TestParseSingleQuotesWithDoubleQuotesInside(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env")
	os.WriteFile(path, []byte("FOO='say \"hello\"'\n"), 0644)

	vars, err := Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vars["FOO"] != `say "hello"` {
		t.Errorf("expected 'say \"hello\"', got %q", vars["FOO"])
	}
}

func TestParseDoubleQuotesWithSingleQuotesInside(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env")
	os.WriteFile(path, []byte(`FOO="it's working"`+"\n"), 0644)

	vars, err := Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vars["FOO"] != "it's working" {
		t.Errorf("expected \"it's working\", got %q", vars["FOO"])
	}
}

func TestParseMultilineValue(t *testing.T) {
	// Standard .env format doesn't support multiline, but quoted strings with \n do
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env")
	os.WriteFile(path, []byte("FOO=\"line1\\nline2\\nline3\"\n"), 0644)

	vars, err := Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "line1\nline2\nline3"
	if vars["FOO"] != expected {
		t.Errorf("expected %q, got %q", expected, vars["FOO"])
	}
}

func TestParseSpecialCharactersInValue(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env")
	os.WriteFile(path, []byte("FOO=$BAR\nBAZ=`cmd`\nQUX=!@#\n"), 0644)

	vars, err := Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vars["FOO"] != "$BAR" {
		t.Errorf("expected '$BAR', got %q", vars["FOO"])
	}
	if vars["BAZ"] != "`cmd`" {
		t.Errorf("expected '`cmd`', got %q", vars["BAZ"])
	}
	if vars["QUX"] != "!@#" {
		t.Errorf("expected '!@#', got %q", vars["QUX"])
	}
}

func TestParseDollarSignInValue(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env")
	os.WriteFile(path, []byte("PRICE=$100\n"), 0644)

	vars, err := Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vars["PRICE"] != "$100" {
		t.Errorf("expected '$100', got %q", vars["PRICE"])
	}
}

func TestParseConsecutiveEmptyLines(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env")
	content := "FOO=bar\n\n\n\nBAR=baz\n"
	os.WriteFile(path, []byte(content), 0644)

	vars, err := Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vars) != 2 {
		t.Errorf("expected 2 vars, got %d", len(vars))
	}
}

func TestParseLineEmptyValue(t *testing.T) {
	key, val, err := parseLine("FOO=")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "FOO" || val != "" {
		t.Errorf("expected FOO='', got %s=%q", key, val)
	}
}

func TestParseLineOnlySpaces(t *testing.T) {
	// A line with just spaces should be treated as empty and skipped
	key, val, err := parseLine("   ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "" || val != "" {
		t.Errorf("expected empty key/val for whitespace-only line, got %s=%q", key, val)
	}
}

func TestParseKeyWithSpaces(t *testing.T) {
	// Key with spaces around it should be trimmed
	key, val, err := parseLine("  FOO  =bar")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "FOO" || val != "bar" {
		t.Errorf("expected FOO=bar, got %s=%q", key, val)
	}
}

func TestUnquoteIncompleteEscape(t *testing.T) {
	// "hello\ is a malformed string (unterminated quote)
	// unquote should return it as-is since closing quote is missing
	got := unquote(`"hello\`)
	expected := `"hello\`
	if got != expected {
		t.Errorf("unquote(\"hello\\\") = %q, want %q", got, expected)
	}
}

func TestParseValueWithTabs(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env")
	os.WriteFile(path, []byte("FOO=\thello\tworld\t\n"), 0644)

	vars, err := Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// TrimSpace removes tabs too
	if vars["FOO"] != "hello\tworld" {
		t.Errorf("expected 'hello\\tworld' (trimmed), got %q", vars["FOO"])
	}
}

func TestParseWindowsPath(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env")
	os.WriteFile(path, []byte(`DATA_DIR=C:\Users\Admin\Data`+"\n"), 0644)

	vars, err := Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(vars["DATA_DIR"], `\`) {
		t.Errorf("expected Windows path with backslashes, got %q", vars["DATA_DIR"])
	}
}

func TestParseVeryLongValue(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env")
	longValue := strings.Repeat("a", 100000)
	os.WriteFile(path, []byte("FOO="+longValue+"\n"), 0644)

	vars, err := Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vars["FOO"]) != 100000 {
		t.Errorf("expected length 100000, got %d", len(vars["FOO"]))
	}
}

func TestParseManyVariables(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env")
	var sb strings.Builder
	for i := 0; i < 1000; i++ {
		sb.WriteString(fmt.Sprintf("VAR_%04d=value\n", i))
	}
	os.WriteFile(path, []byte(sb.String()), 0644)

	vars, err := Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vars) != 1000 {
		t.Errorf("expected 1000 vars, got %d", len(vars))
	}
}
