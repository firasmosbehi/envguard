// Package cli tests for scan, generate, and validate commands.
package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// === Scan command tests ===

func TestScanCommandWithSecrets(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	envContent := "AWS_KEY=AKIAIOSFODNN7EXAMPLE\nGITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\nPORT=3000\n"
	os.WriteFile(envPath, []byte(envContent), 0644)

	cmd := newScanCmd()
	cmd.SetArgs([]string{"--env", envPath})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for secrets found")
	}
	out := buf.String()
	if !strings.Contains(out, "aws-access-key") {
		t.Errorf("expected aws-access-key detection, got: %s", out)
	}
	if !strings.Contains(out, "github-token") {
		t.Errorf("expected github-token detection, got: %s", out)
	}
}

func TestScanCommandClean(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	os.WriteFile(envPath, []byte("PORT=3000\nHOST=localhost\n"), 0644)

	cmd := newScanCmd()
	cmd.SetArgs([]string{"--env", envPath})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "No secrets detected") {
		t.Errorf("expected clean scan, got: %s", buf.String())
	}
}

func TestScanCommandJSON(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	os.WriteFile(envPath, []byte("AWS_KEY=AKIAIOSFODNN7EXAMPLE\n"), 0644)

	cmd := newScanCmd()
	cmd.SetArgs([]string{"--env", envPath, "--format", "json"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for secrets found")
	}
	out := buf.String()
	if !strings.Contains(out, `"key"`) {
		t.Errorf("expected JSON output, got: %s", out)
	}
}

func TestScanCommandUnknownFormat(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	// Include a secret so the scan reaches the format switch
	os.WriteFile(envPath, []byte("AWS_KEY=AKIAIOSFODNN7EXAMPLE\n"), 0644)

	cmd := newScanCmd()
	cmd.SetArgs([]string{"--env", envPath, "--format", "xml"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
	if !strings.Contains(buf.String(), "unknown format") {
		t.Errorf("expected unknown format error, got: %s", buf.String())
	}
}

func TestScanCommandMissingEnvFile(t *testing.T) {
	cmd := newScanCmd()
	cmd.SetArgs([]string{"--env", "/nonexistent/.env"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing env file")
	}
	if !strings.Contains(buf.String(), "failed to open env file") {
		t.Errorf("expected env file error, got: %s", buf.String())
	}
}

func TestScanCommandWithCustomRulesSchema(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")

	os.WriteFile(envPath, []byte("CUSTOM_TOKEN=iat_abc123def456ghi789jkl012mno345pq\n"), 0644)
	os.WriteFile(schemaPath, []byte(`
version: "1.0"
env:
  CUSTOM_TOKEN:
    type: string
secrets:
  custom:
    - name: internal-api-token
      pattern: "iat_[a-zA-Z0-9]{32}"
      message: "Internal API token detected"
`), 0644)

	cmd := newScanCmd()
	cmd.SetArgs([]string{"--env", envPath, "--schema", schemaPath})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for custom secret found")
	}
	if !strings.Contains(buf.String(), "internal-api-token") {
		t.Errorf("expected custom rule detection, got: %s", buf.String())
	}
}

// === Generate-example command tests ===

func TestGenerateExampleCommand(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	outputPath := filepath.Join(tmpDir, ".env.example")

	schemaContent := `
version: "1.0"
env:
  DATABASE_URL:
    type: string
    required: true
    description: "Database connection string"
  PORT:
    type: integer
    default: 3000
  DEBUG:
    type: boolean
    default: false
`
	os.WriteFile(schemaPath, []byte(schemaContent), 0644)

	cmd := newGenerateExampleCmd()
	cmd.SetArgs([]string{"--schema", schemaPath, "--output", outputPath})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("expected output file to exist: %v", err)
	}

	if !strings.Contains(string(content), "DATABASE_URL=") {
		t.Errorf("expected DATABASE_URL in output, got: %s", string(content))
	}
	if !strings.Contains(string(content), "PORT=") {
		t.Errorf("expected PORT in output, got: %s", string(content))
	}
	if !strings.Contains(string(content), "DEBUG=") {
		t.Errorf("expected DEBUG in output, got: %s", string(content))
	}
}

func TestGenerateExampleCommandFileExists(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	outputPath := filepath.Join(tmpDir, ".env.example")

	schemaContent := `version: "1.0"
env:
  FOO:
    type: string
`
	os.WriteFile(schemaPath, []byte(schemaContent), 0644)
	os.WriteFile(outputPath, []byte("exists\n"), 0644)

	cmd := newGenerateExampleCmd()
	cmd.SetArgs([]string{"--schema", schemaPath, "--output", outputPath})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when output file already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected already exists error, got: %v", err)
	}
}

func TestGenerateExampleCommandMissingSchema(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "nonexistent.yaml")
	outputPath := filepath.Join(tmpDir, ".env.example")

	cmd := newGenerateExampleCmd()
	cmd.SetArgs([]string{"--schema", schemaPath, "--output", outputPath})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing schema")
	}
}

// === Validate command edge cases ===

func TestValidateCommandMultipleEnvFiles(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath1 := filepath.Join(tmpDir, ".env")
	envPath2 := filepath.Join(tmpDir, ".env.local")

	schemaContent := `
version: "1.0"
env:
  SHARED:
    type: string
  LOCAL:
    type: string
  OVERRIDE:
    type: string
`
	os.WriteFile(schemaPath, []byte(schemaContent), 0644)
	os.WriteFile(envPath1, []byte("SHARED=from-base\nOVERRIDE=base-value\n"), 0644)
	os.WriteFile(envPath2, []byte("LOCAL=from-local\nOVERRIDE=local-value\n"), 0644)

	cmd := newValidateCmd()
	cmd.SetArgs([]string{"--schema", schemaPath, "--env", envPath1, "--env", envPath2})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "validated") {
		t.Errorf("expected success, got: %s", buf.String())
	}
}

func TestValidateCommandScanSecrets(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")

	os.WriteFile(schemaPath, []byte(`
version: "1.0"
env:
  API_KEY:
    type: string
    required: true
`), 0644)
	os.WriteFile(envPath, []byte("API_KEY=AKIAIOSFODNN7EXAMPLE\n"), 0644)

	cmd := newValidateCmd()
	cmd.SetArgs([]string{"--schema", schemaPath, "--env", envPath, "--scan-secrets"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for secret detected")
	}
	if !strings.Contains(buf.String(), "secret") {
		t.Errorf("expected secret error, got: %s", buf.String())
	}
}

func TestValidateCommandGitHubFormat(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")

	os.WriteFile(schemaPath, []byte(`
version: "1.0"
env:
  FOO:
    type: string
    required: true
`), 0644)
	os.WriteFile(envPath, []byte("FOO=bar\n"), 0644)

	cmd := newValidateCmd()
	cmd.SetArgs([]string{"--schema", schemaPath, "--env", envPath, "--format", "github"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateCommandEnvName(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")

	os.WriteFile(schemaPath, []byte(`
version: "1.0"
env:
  DB_PASSWORD:
    type: string
    requiredIn: ["production"]
`), 0644)

	cmd := newValidateCmd()
	cmd.SetArgs([]string{"--schema", schemaPath, "--env", envPath, "--env-name", "production"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing requiredIn var in production")
	}
}

func TestValidateCommandMissingEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")

	os.WriteFile(schemaPath, []byte(`
version: "1.0"
env:
  FOO:
    type: string
`), 0644)

	cmd := newValidateCmd()
	cmd.SetArgs([]string{"--schema", schemaPath, "--env", "/nonexistent/.env"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing env file")
	}
}

// === Root command / Execute tests ===

func TestExecuteVersion(t *testing.T) {
	// Use a fresh root to avoid global state pollution
	localRoot := &cobra.Command{
		Use:   "envguard",
		Short: "Test root",
	}
	localRoot.AddCommand(newVersionCmd("0.1.0-test"))
	localRoot.SetArgs([]string{"version"})
	var buf bytes.Buffer
	localRoot.SetOut(&buf)
	localRoot.SetErr(&buf)

	if err := localRoot.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "0.1.0-test") {
		t.Errorf("expected version output, got: %s", buf.String())
	}
}
