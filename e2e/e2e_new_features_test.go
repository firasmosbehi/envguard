package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestE2E_ScanCommand(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	env := "AWS_KEY=AKIAIOSFODNN7EXAMPLE\nGITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\nPORT=3000\n"

	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "scan", "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "aws-access-key") {
		t.Errorf("expected aws-access-key detection, got: %s", out)
	}
	if !strings.Contains(out, "github-token") {
		t.Errorf("expected github-token detection, got: %s", out)
	}
}

func TestE2E_ScanCommandClean(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	env := "PORT=3000\nHOST=localhost\n"

	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "scan", "--env", envPath)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "No secrets detected") {
		t.Errorf("expected clean scan, got: %s", out)
	}
}

func TestE2E_ValidateWithSecretScan(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  API_KEY:
    type: string
    required: true
`
	env := "API_KEY=AKIAIOSFODNN7EXAMPLE\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath, "--scan-secrets")
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "secret") {
		t.Errorf("expected secret scan error, got: %s", out)
	}
}

func TestE2E_FormatBase64(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  ENCODED:
    type: string
    format: base64
`
	env := "ENCODED=aGVsbG8=\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "validated") {
		t.Errorf("expected success, got: %s", out)
	}
}

func TestE2E_FormatIP(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  SERVER_IP:
    type: string
    format: ip
`
	env := "SERVER_IP=192.168.1.1\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %s", code, out)
	}
}

func TestE2E_FormatPort(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  PORT:
    type: string
    format: port
`
	env := "PORT=8080\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %s", code, out)
	}
}

func TestE2E_FormatJSON(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  CONFIG:
    type: string
    format: json
`
	env := "CONFIG={\"key\":\"value\"}\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %s", code, out)
	}
}

func TestE2E_SchemaInheritance(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	baseSchema := `
version: "1.0"
env:
  DATABASE_URL:
    type: string
    required: true
  PORT:
    type: integer
    default: 3000
`
	extendedSchema := `
version: "1.0"
extends: ./base.yaml
env:
  API_KEY:
    type: string
    required: true
`

	basePath := filepath.Join(tmpDir, "base.yaml")
	extendedPath := filepath.Join(tmpDir, "extended.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(basePath, []byte(baseSchema), 0644)
	os.WriteFile(extendedPath, []byte(extendedSchema), 0644)
	os.WriteFile(envPath, []byte("DATABASE_URL=postgres://localhost\nAPI_KEY=secret\n"), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", extendedPath, "--env", envPath)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "validated") {
		t.Errorf("expected success, got: %s", out)
	}
}
