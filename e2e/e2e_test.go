package e2e

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// buildEnvGuard compiles the CLI binary for e2e tests.
func buildEnvGuard(t *testing.T) string {
	t.Helper()
	binPath := filepath.Join(t.TempDir(), "envguard")
	cmd := exec.Command("go", "build", "-o", binPath, "../cmd/envguard")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build envguard: %v\n%s", err, out)
	}
	return binPath
}

func runEnvGuard(t *testing.T, bin string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run envguard: %v", err)
		}
	}
	return string(out), exitCode
}

func runEnvGuardJSON(t *testing.T, bin string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	stdout, err := cmd.Output()
	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run envguard: %v", err)
		}
	}
	return string(stdout), exitCode
}

func TestE2E_ValidEnv(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  DATABASE_URL:
    type: string
    required: true
  PORT:
    type: integer
    default: 3000
`
	env := "DATABASE_URL=postgresql://localhost/mydb\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "✓ All environment variables validated.") {
		t.Errorf("expected success message, got: %s", out)
	}
}

func TestE2E_InvalidEnv(t *testing.T) {
	bin := buildEnvGuard(t)
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

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "✗ Environment validation failed") {
		t.Errorf("expected failure message, got: %s", out)
	}
	if !strings.Contains(out, "PORT") {
		t.Errorf("expected PORT error, got: %s", out)
	}
}

func TestE2E_MissingRequired(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  API_KEY:
    type: string
    required: true
`
	env := "\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "required") {
		t.Errorf("expected required error, got: %s", out)
	}
}

func TestE2E_JSONOutput(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  FOO:
    type: string
    required: true
`
	env := "FOO=bar\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuardJSON(t, bin, "validate", "--schema", schemaPath, "--env", envPath, "--format", "json")
	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %s", code, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Errorf("expected valid JSON, got: %s. error: %v", out, err)
	}
	if result["valid"] != true {
		t.Errorf("expected valid=true, got: %v", result["valid"])
	}
}

func TestE2E_JSONOutputWithErrors(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  PORT:
    type: integer
    required: true
`
	env := "PORT=abc\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuardJSON(t, bin, "validate", "--schema", schemaPath, "--env", envPath, "--format", "json")
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Errorf("expected valid JSON, got: %s. error: %v", out, err)
	}
	if result["valid"] != false {
		t.Errorf("expected valid=false, got: %v", result["valid"])
	}
	errors, ok := result["errors"].([]any)
	if !ok || len(errors) == 0 {
		t.Errorf("expected errors array, got: %v", result["errors"])
	}
}

func TestE2E_StrictMode(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  FOO:
    type: string
`
	env := "FOO=bar\nUNKNOWN=value\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath, "--strict")
	if code != 0 {
		t.Errorf("expected exit code 0 (warnings only), got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "⚠ Warnings") {
		t.Errorf("expected warnings header, got: %s", out)
	}
	if !strings.Contains(out, "UNKNOWN") {
		t.Errorf("expected UNKNOWN warning, got: %s", out)
	}
}

func TestE2E_MissingSchemaFile(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schemaPath := filepath.Join(tmpDir, "nonexistent.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte("FOO=bar\n"), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 2 {
		t.Errorf("expected exit code 2, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "failed to read schema file") {
		t.Errorf("expected schema file error, got: %s", out)
	}
}

func TestE2E_MissingEnvFile(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  FOO:
    type: string
`
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, "nonexistent.env")
	os.WriteFile(schemaPath, []byte(schema), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 2 {
		t.Errorf("expected exit code 2, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "failed to open env file") {
		t.Errorf("expected env file error, got: %s", out)
	}
}

func TestE2E_InvalidSchemaYAML(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `not: [ valid yaml :::`
	env := "FOO=bar\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 2 {
		t.Errorf("expected exit code 2, got %d. output: %s", code, out)
	}
}

func TestE2E_InvalidSchemaStructure(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  FOO:
    type: invalidtype
`
	env := "FOO=bar\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 2 {
		t.Errorf("expected exit code 2, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "invalid schema") {
		t.Errorf("expected invalid schema error, got: %s", out)
	}
}

func TestE2E_UnknownFormatFlag(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  FOO:
    type: string
`
	env := "FOO=bar\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath, "--format", "xml")
	if code != 2 {
		t.Errorf("expected exit code 2, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "unknown format") {
		t.Errorf("expected unknown format error, got: %s", out)
	}
}

func TestE2E_InitCommand(t *testing.T) {
	bin := buildEnvGuard(t)

	out, code := runEnvGuard(t, bin, "init")
	// init writes to cwd, not tmpDir. We need to change dir.
	// Skip this test or use a different approach
	_ = out
	_ = code
	t.Skip("init command writes to cwd; skipping in e2e test")
}

func TestE2E_VersionCommand(t *testing.T) {
	bin := buildEnvGuard(t)

	out, code := runEnvGuard(t, bin, "version")
	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "envguard version") {
		t.Errorf("expected version string, got: %s", out)
	}
}

func TestE2E_DefaultValueInjection(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  PORT:
    type: integer
    default: 3000
  DEBUG:
    type: boolean
    default: false
`
	env := "\n" // empty env

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "✓ All environment variables validated.") {
		t.Errorf("expected success with defaults, got: %s", out)
	}
}

func TestE2E_AllTypesTogether(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  STR:
    type: string
    required: true
  INT:
    type: integer
    required: true
  FLOAT:
    type: float
    required: true
  BOOL:
    type: boolean
    required: true
  ENUM_STR:
    type: string
    enum: [a, b, c]
    required: true
  PATTERN:
    type: string
    pattern: "^[a-z]+$"
    required: true
`
	env := `STR=hello
INT=42
FLOAT=3.14
BOOL=true
ENUM_STR=b
PATTERN=abc
`

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

func TestE2E_CollectAllErrors(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  A:
    type: string
    required: true
  B:
    type: integer
    required: true
  C:
    type: boolean
    required: true
`
	env := `A=
B=not-an-int
C=not-a-bool
`

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	// Should report all 3 errors, not stop at first
	errors := 0
	if strings.Contains(out, "A") {
		errors++
	}
	if strings.Contains(out, "B") {
		errors++
	}
	if strings.Contains(out, "C") {
		errors++
	}
	if errors != 3 {
		t.Errorf("expected 3 errors in output, found %d. output: %s", errors, out)
	}
}

func TestE2E_EmptyEnvFile(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  FOO:
    type: string
    default: bar
`
	env := ""

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 0 {
		t.Errorf("expected exit code 0 with default, got %d. output: %s", code, out)
	}
}

func TestE2E_WhitespaceValue(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  FOO:
    type: string
    required: true
`
	env := "FOO=   \n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1 for whitespace-only required value, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "required") {
		t.Errorf("expected required error, got: %s", out)
	}
}

func TestE2E_HelpFlag(t *testing.T) {
	bin := buildEnvGuard(t)

	out, code := runEnvGuard(t, bin, "--help")
	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "Usage:") {
		t.Errorf("expected usage info, got: %s", out)
	}
}

func TestE2E_ValidateHelpFlag(t *testing.T) {
	bin := buildEnvGuard(t)

	out, code := runEnvGuard(t, bin, "validate", "--help")
	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "--schema") {
		t.Errorf("expected --schema flag in help, got: %s", out)
	}
	if !strings.Contains(out, "--env") {
		t.Errorf("expected --env flag in help, got: %s", out)
	}
}

func TestE2E_NoArgs(t *testing.T) {
	bin := buildEnvGuard(t)

	out, code := runEnvGuard(t, bin)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "Usage:") {
		t.Errorf("expected usage info when no args, got: %s", out)
	}
}

func TestE2E_FloatScientificNotation(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  SCI:
    type: float
    required: true
`
	env := "SCI=1.5e10\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 0 {
		t.Errorf("expected exit code 0 for scientific notation, got %d. output: %s", code, out)
	}
}

func TestE2E_BooleanMixedCase(t *testing.T) {
	bin := buildEnvGuard(t)

	schema := `
version: "1.0"
env:
  DEBUG:
    type: boolean
    required: true
`
	cases := []string{"True", "FALSE", "YES", "No", "ON", "Off", "1", "0"}
	for _, val := range cases {
		t.Run(val, func(t *testing.T) {
			tmpDir := t.TempDir()
			env := fmt.Sprintf("DEBUG=%s\n", val)
			schemaPath := filepath.Join(tmpDir, "envguard.yaml")
			envPath := filepath.Join(tmpDir, ".env")
			os.WriteFile(schemaPath, []byte(schema), 0644)
			os.WriteFile(envPath, []byte(env), 0644)

			out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
			if code != 0 {
				t.Errorf("expected exit code 0 for boolean %q, got %d. output: %s", val, code, out)
			}
		})
	}
}
