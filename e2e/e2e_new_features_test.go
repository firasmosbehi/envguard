package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestE2E_MinMaxInteger(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  PORT:
    type: integer
    min: 1024
    max: 65535
`
	env := "PORT=80\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "min") {
		t.Errorf("expected min error, got: %s", out)
	}
}

func TestE2E_MinMaxFloat(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  RATIO:
    type: float
    min: 0.0
    max: 1.0
`
	env := "RATIO=1.5\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "max") {
		t.Errorf("expected max error, got: %s", out)
	}
}

func TestE2E_StringLength(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  TOKEN:
    type: string
    minLength: 8
    maxLength: 32
`
	env := "TOKEN=short\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "minLength") {
		t.Errorf("expected minLength error, got: %s", out)
	}
}

func TestE2E_FormatEmail(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  EMAIL:
    type: string
    format: email
`
	env := "EMAIL=invalid-email\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "format") {
		t.Errorf("expected format error, got: %s", out)
	}
}

func TestE2E_FormatURL(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  ENDPOINT:
    type: string
    format: url
`
	env := "ENDPOINT=https://api.example.com\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %s", code, out)
	}
}

func TestE2E_Disallow(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  API_KEY:
    type: string
    disallow: ["undefined", "null", ""]
`
	env := "API_KEY=undefined\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "disallow") {
		t.Errorf("expected disallow error, got: %s", out)
	}
}

func TestE2E_EnvNameRequiredIn(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  DB_URL:
    type: string
    requiredIn: ["production", "staging"]
`
	env := "DB_URL=\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	// Should fail in production
	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath, "--env-name", "production")
	if code != 1 {
		t.Errorf("expected exit code 1 in production, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "required") {
		t.Errorf("expected required error, got: %s", out)
	}

	// Should pass in development
	out, code = runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath, "--env-name", "development")
	if code != 0 {
		t.Errorf("expected exit code 0 in development, got %d. output: %s", code, out)
	}
}

func TestE2E_EnvNameDevOnly(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  DEBUG_TOOL:
    type: string
    devOnly: true
`
	env := "DEBUG_TOOL=\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	// Should be ignored in production
	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath, "--env-name", "production")
	if code != 0 {
		t.Errorf("expected exit code 0 in production, got %d. output: %s", code, out)
	}

	// Should fail in development
	out, code = runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath, "--env-name", "development")
	if code != 1 {
		t.Errorf("expected exit code 1 in development, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "required") {
		t.Errorf("expected required error, got: %s", out)
	}
}

func TestE2E_GenerateExample(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
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
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte(schema), 0644)

	examplePath := filepath.Join(tmpDir, ".env.example")
	out, code := runEnvGuard(t, bin, "generate-example", "--schema", schemaPath, "--output", examplePath)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %s", code, out)
	}

	content, err := os.ReadFile(examplePath)
	if err != nil {
		t.Fatalf("expected .env.example to be created: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "DATABASE_URL=") {
		t.Errorf("expected DATABASE_URL in example, got: %s", contentStr)
	}
	if !strings.Contains(contentStr, "PORT=3000") {
		t.Errorf("expected PORT=3000 in example, got: %s", contentStr)
	}
	if !strings.Contains(contentStr, "DEBUG=false") {
		t.Errorf("expected DEBUG=false in example, got: %s", contentStr)
	}
}
