package envguard

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateFile(t *testing.T) {
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  API_KEY:
    type: string
    required: true
  PORT:
    type: integer
    default: 3000
`
	env := "API_KEY=secret123\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	result, err := ValidateFile(schemaPath, envPath, false, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid, got errors: %v", result.Errors)
	}
}

func TestValidateFileInvalid(t *testing.T) {
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  PORT:
    type: integer
    required: true
`
	env := "PORT=not-a-number\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	result, err := ValidateFile(schemaPath, envPath, false, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Valid {
		t.Error("expected invalid")
	}
	if len(result.Errors) == 0 {
		t.Error("expected at least one error")
	}
}

func TestParseSchema(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte(`
version: "1.0"
env:
  FOO:
    type: string
`), 0644)

	s, err := ParseSchema(schemaPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Version != "1.0" {
		t.Errorf("version = %q, want 1.0", s.Version)
	}
	if _, ok := s.Env["FOO"]; !ok {
		t.Error("FOO should be in schema")
	}
}

func TestParseEnv(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte("FOO=bar\nBAZ=qux\n"), 0644)

	vars, err := ParseEnv(envPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vars["FOO"] != "bar" {
		t.Errorf("FOO = %q, want bar", vars["FOO"])
	}
	if vars["BAZ"] != "qux" {
		t.Errorf("BAZ = %q, want qux", vars["BAZ"])
	}
}
