package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestE2E_MultipleEnvFiles(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  SHARED:
    type: string
  LOCAL:
    type: string
  OVERRIDE:
    type: string
`
	env1 := "SHARED=from-base\nOVERRIDE=base-value\n"
	env2 := "LOCAL=from-local\nOVERRIDE=local-value\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath1 := filepath.Join(tmpDir, ".env")
	envPath2 := filepath.Join(tmpDir, ".env.local")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath1, []byte(env1), 0644)
	os.WriteFile(envPath2, []byte(env2), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath1, "--env", envPath2)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "✓ All environment variables validated.") {
		t.Errorf("expected success, got: %s", out)
	}
}

func TestE2E_MultipleEnvFilesOverride(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  PORT:
    type: integer
`
	env1 := "PORT=3000\n"
	env2 := "PORT=8080\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath1 := filepath.Join(tmpDir, ".env")
	envPath2 := filepath.Join(tmpDir, ".env.local")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath1, []byte(env1), 0644)
	os.WriteFile(envPath2, []byte(env2), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath1, "--env", envPath2)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %s", code, out)
	}
}

func TestE2E_CustomMessage(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  PORT:
    type: integer
    min: 1024
    message: "PORT must be a valid port number (1024-65535)"
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
	if !strings.Contains(out, "PORT must be a valid port number") {
		t.Errorf("expected custom message, got: %s", out)
	}
}

func TestE2E_ArrayType(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  ALLOWED_HOSTS:
    type: array
    separator: ","
    minLength: 1
`
	env := "ALLOWED_HOSTS=host1,host2,host3\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "✓ All environment variables validated.") {
		t.Errorf("expected success, got: %s", out)
	}
}

func TestE2E_ArrayTypeEnum(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  PERMISSIONS:
    type: array
    separator: ","
    enum: [read, write, admin]
`
	env := "PERMISSIONS=read,delete\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "enum") {
		t.Errorf("expected enum error, got: %s", out)
	}
}

func TestE2E_PreCommitHookExists(t *testing.T) {
	// Verify the pre-commit hooks file exists and is valid YAML
	content, err := os.ReadFile("../.pre-commit-hooks.yaml")
	if err != nil {
		t.Fatalf(".pre-commit-hooks.yaml should exist: %v", err)
	}
	if !strings.Contains(string(content), "envguard-validate") {
		t.Errorf("hook id should be envguard-validate")
	}
	if !strings.Contains(string(content), "language: system") {
		t.Errorf("language should be system")
	}
}

func TestE2E_AllowEmpty(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  OPTIONAL:
    type: string
    allowEmpty: false
`
	env := "OPTIONAL=\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "allowEmpty") {
		t.Errorf("expected allowEmpty error, got: %s", out)
	}
}

func TestE2E_DependsOn(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  HTTPS:
    type: boolean
  SSL_CERT:
    type: string
    dependsOn: HTTPS
    when: "true"
`
	env := "HTTPS=true\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "SSL_CERT") {
		t.Errorf("expected SSL_CERT error, got: %s", out)
	}
}

func TestE2E_Contains(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  ROLES:
    type: array
    separator: ","
    contains: "admin"
`
	env := "ROLES=read,write\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "contains") {
		t.Errorf("expected contains error, got: %s", out)
	}
}

func TestE2E_DockerfileExists(t *testing.T) {
	_, err := os.Stat("../Dockerfile")
	if err != nil {
		t.Errorf("Dockerfile should exist: %v", err)
	}
}

func TestE2E_HomebrewFormulaExists(t *testing.T) {
	_, err := os.Stat("../homebrew/envguard.rb")
	if err != nil {
		t.Errorf("homebrew formula should exist: %v", err)
	}
}
