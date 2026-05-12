package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLintCommandValidSchema(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")

	schemaContent := `
version: "1.0"
env:
  FOO:
    type: string
    required: true
    description: "A required variable"
`
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("failed to write schema: %v", err)
	}

	cmd := newLintCmd()
	cmd.SetArgs([]string{"--schema", schemaPath})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "✓ Schema passes all lint checks.") {
		t.Errorf("expected success message, got: %s", buf.String())
	}
}

func TestLintCommandWithIssues(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")

	schemaContent := `
version: "1.0"
env:
  FOO:
    type: string
    required: true
    default: "bar"
  BAR:
    type: string
    required: true
    allowEmpty: false
  BAZ:
    type: string
    dependsOn: UNKNOWN
    when: "true"
`
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("failed to write schema: %v", err)
	}

	cmd := newLintCmd()
	cmd.SetArgs([]string{"--schema", schemaPath})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for schema with issues")
	}
	out := buf.String()
	if !strings.Contains(out, "✗ Schema lint found") {
		t.Errorf("expected failure header, got: %s", out)
	}
	if !strings.Contains(out, "required and default are mutually exclusive") {
		t.Errorf("expected redundant rule message, got: %s", out)
	}
	if !strings.Contains(out, "dependsOn references undefined variable") {
		t.Errorf("expected unreachable dependency message, got: %s", out)
	}
}

func TestLintCommandJSONFormat(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")

	schemaContent := `
version: "1.0"
env:
  FOO:
    type: string
    required: true
    default: "bar"
`
	os.WriteFile(schemaPath, []byte(schemaContent), 0644)

	cmd := newLintCmd()
	cmd.SetArgs([]string{"--schema", schemaPath, "--format", "json"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for schema with issues")
	}
	out := buf.String()
	if !strings.Contains(out, `"level"`) {
		t.Errorf("expected JSON output, got: %s", out)
	}
}
