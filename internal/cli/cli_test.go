package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateCommand(t *testing.T) {
	// Create a temporary schema and .env
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")

	schemaContent := `
version: "1.0"
env:
  FOO:
    type: string
    required: true
`
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("failed to write schema: %v", err)
	}

	// Valid env
	if err := os.WriteFile(envPath, []byte("FOO=bar\n"), 0644); err != nil {
		t.Fatalf("failed to write env: %v", err)
	}

	cmd := newValidateCmd()
	cmd.SetArgs([]string{"--schema", schemaPath, "--env", envPath})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("valid env should not error: %v", err)
	}
	if !strings.Contains(buf.String(), "✓ All environment variables validated.") {
		t.Errorf("expected success message, got: %s", buf.String())
	}

	// Invalid env
	if err := os.WriteFile(envPath, []byte("FOO=\n"), 0644); err != nil {
		t.Fatalf("failed to write env: %v", err)
	}

	cmd = newValidateCmd()
	cmd.SetArgs([]string{"--schema", schemaPath, "--env", envPath})
	buf.Reset()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid env")
	}
	if !strings.Contains(buf.String(), "✗ Environment validation failed") {
		t.Errorf("expected failure message, got: %s", buf.String())
	}
}

func TestValidateCommandJSON(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")

	schemaContent := `
version: "1.0"
env:
  FOO:
    type: string
    required: true
`
	os.WriteFile(schemaPath, []byte(schemaContent), 0644)
	os.WriteFile(envPath, []byte("FOO=bar\n"), 0644)

	cmd := newValidateCmd()
	cmd.SetArgs([]string{"--schema", schemaPath, "--env", envPath, "--format", "json"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), `"valid": true`) {
		t.Errorf("expected JSON with valid=true, got: %s", buf.String())
	}
}

func TestInitCommand(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	cmd := newInitCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "Created envguard.yaml") {
		t.Errorf("expected creation message, got: %s", buf.String())
	}

	// Second init should fail
	cmd = newInitCmd()
	buf.Reset()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when file already exists")
	}
}

func TestVersionCommand(t *testing.T) {
	cmd := newVersionCmd("0.1.0")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "envguard version 0.1.0") {
		t.Errorf("expected version string, got: %s", buf.String())
	}
}
