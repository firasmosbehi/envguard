package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestE2E_ScanCommandJSON(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	env := "AWS_KEY=AKIAIOSFODNN7EXAMPLE\n"
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "scan", "--env", envPath, "--format", "json")
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, `"key"`) {
		t.Errorf("expected JSON output, got: %s", out)
	}
}

func TestE2E_ScanCommandNoSecrets(t *testing.T) {
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
		t.Errorf("expected clean scan message, got: %s", out)
	}
}

func TestE2E_ScanCommandMissingEnvFile(t *testing.T) {
	bin := buildEnvGuard(t)

	out, code := runEnvGuard(t, bin, "scan", "--env", "/nonexistent/.env")
	if code != 2 {
		t.Errorf("expected exit code 2, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "failed to open env file") {
		t.Errorf("expected env file error, got: %s", out)
	}
}

func TestE2E_ScanCommandUnknownFormat(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	envPath := filepath.Join(tmpDir, ".env")
	// Include a secret so scan reaches the format switch
	os.WriteFile(envPath, []byte("AWS_KEY=AKIAIOSFODNN7EXAMPLE\n"), 0644)

	out, code := runEnvGuard(t, bin, "scan", "--env", envPath, "--format", "xml")
	if code != 2 {
		t.Errorf("expected exit code 2, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "unknown format") {
		t.Errorf("expected unknown format error, got: %s", out)
	}
}

func TestE2E_LintCommandValid(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  FOO:
    type: string
    required: true
    description: "A required variable"
`
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte(schema), 0644)

	out, code := runEnvGuard(t, bin, "lint", "--schema", schemaPath)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "passes all lint checks") {
		t.Errorf("expected success message, got: %s", out)
	}
}

func TestE2E_LintCommandWithIssues(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  FOO:
    type: string
    required: true
    default: "bar"
  BAR:
    type: string
    dependsOn: UNKNOWN
    when: "true"
`
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte(schema), 0644)

	out, code := runEnvGuard(t, bin, "lint", "--schema", schemaPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "required and default are mutually exclusive") {
		t.Errorf("expected redundant rule message, got: %s", out)
	}
	if !strings.Contains(out, "dependsOn references undefined variable") {
		t.Errorf("expected unreachable dependency message, got: %s", out)
	}
}

func TestE2E_LintCommandJSONFormat(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  FOO:
    type: string
    required: true
    default: "bar"
`
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte(schema), 0644)

	out, code := runEnvGuard(t, bin, "lint", "--schema", schemaPath, "--format", "json")
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, `"level"`) {
		t.Errorf("expected JSON output, got: %s", out)
	}
}

func TestE2E_LintCommandMissingSchema(t *testing.T) {
	bin := buildEnvGuard(t)

	out, code := runEnvGuard(t, bin, "lint", "--schema", "/nonexistent.yaml")
	if code != 2 {
		t.Errorf("expected exit code 2, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "failed to read schema file") {
		t.Errorf("expected schema file error, got: %s", out)
	}
}

func TestE2E_GenerateExampleCommand(t *testing.T) {
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
	outputPath := filepath.Join(tmpDir, ".env.example")
	os.WriteFile(schemaPath, []byte(schema), 0644)

	out, code := runEnvGuard(t, bin, "generate-example", "--schema", schemaPath, "--output", outputPath)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "Created") {
		t.Errorf("expected creation message, got: %s", out)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("expected output file to exist: %v", err)
	}
	if !strings.Contains(string(content), "DATABASE_URL=") {
		t.Errorf("expected DATABASE_URL in output, got: %s", string(content))
	}
}

func TestE2E_GenerateExampleCommandFileExists(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	outputPath := filepath.Join(tmpDir, ".env.example")
	schemaContent := "version: \"1.0\"\nenv:\n  FOO:\n    type: string\n"
	os.WriteFile(schemaPath, []byte(schemaContent), 0644)
	os.WriteFile(outputPath, []byte("exists\n"), 0644)

	out, code := runEnvGuard(t, bin, "generate-example", "--schema", schemaPath, "--output", outputPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "already exists") {
		t.Errorf("expected already exists error, got: %s", out)
	}
}

func TestE2E_GenerateExampleCommandMissingSchema(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	outputPath := filepath.Join(tmpDir, ".env.example")

	out, code := runEnvGuard(t, bin, "generate-example", "--schema", filepath.Join(tmpDir, "nonexistent.yaml"), "--output", outputPath)
	if code != 2 {
		t.Errorf("expected exit code 2, got %d. output: %s", code, out)
	}
}

func TestE2E_VersionCommandStrict(t *testing.T) {
	bin := buildEnvGuard(t)

	out, code := runEnvGuard(t, bin, "version")
	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "envguard version") {
		t.Errorf("expected version string, got: %s", out)
	}
}

func TestE2E_UnknownCommand(t *testing.T) {
	bin := buildEnvGuard(t)

	out, code := runEnvGuard(t, bin, "unknown-cmd")
	if code != 0 {
		// Unknown commands typically show help and exit 0 with cobra
		t.Logf("unknown command exit code: %d, output: %s", code, out)
	}
	if !strings.Contains(out, "Usage:") && !strings.Contains(out, "unknown") {
		t.Errorf("expected usage or unknown command message, got: %s", out)
	}
}

func TestE2E_ValidateWithMultipleEnvFiles(t *testing.T) {
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
	if !strings.Contains(out, "validated") {
		t.Errorf("expected success message, got: %s", out)
	}
}

func TestE2E_ValidateWithEnvNameProduction(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  DB_PASSWORD:
    type: string
    requiredIn: ["production"]
`
	env := ""

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath, "--env-name", "production")
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "required") {
		t.Errorf("expected required error, got: %s", out)
	}
}

func TestE2E_ValidateWithEnvNameDevelopment(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  DB_PASSWORD:
    type: string
    requiredIn: ["production"]
`
	env := ""

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath, "--env-name", "development")
	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %s", code, out)
	}
}

func TestE2E_ValidateWithDevOnly(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  DEBUG_SECRET:
    type: string
    devOnly: true
`
	env := ""

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	// In production, devOnly is skipped
	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath, "--env-name", "production")
	if code != 0 {
		t.Errorf("expected exit code 0 in production, got %d. output: %s", code, out)
	}

	// In development, devOnly is required
	out, code = runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath, "--env-name", "development")
	if code != 1 {
		t.Errorf("expected exit code 1 in development, got %d. output: %s", code, out)
	}
}

func TestE2E_ValidateWithDependsOn(t *testing.T) {
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

func TestE2E_ValidateWithDeprecated(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  OLD_API_URL:
    type: string
    deprecated: "Use API_URL instead"
`
	env := "OLD_API_URL=http://old.example.com\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 0 {
		t.Errorf("expected exit code 0 (warning only), got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "deprecated") {
		t.Errorf("expected deprecated warning, got: %s", out)
	}
}

func TestE2E_ValidateWithSensitive(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  PASSWORD:
    type: string
    required: true
    pattern: "^[a-z]+$"
    sensitive: true
`
	env := "PASSWORD=Secret123\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	// Value should be redacted
	if strings.Contains(out, "Secret123") {
		t.Errorf("expected sensitive value to be redacted, got: %s", out)
	}
	if !strings.Contains(out, "***") {
		t.Errorf("expected redacted placeholder, got: %s", out)
	}
}

func TestE2E_ValidateWithTransform(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  LOWERCASE:
    type: string
    required: true
    transform: lowercase
    enum: [a, b, c]
`
	env := "LOWERCASE=A\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 0 {
		t.Errorf("expected exit code 0 (transform should convert A to a), got %d. output: %s", code, out)
	}
}

func TestE2E_ValidateWithDisallow(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  ENV:
    type: string
    disallow: ["production", "staging"]
`
	env := "ENV=production\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "disallow") || !strings.Contains(out, "not allowed") {
		t.Errorf("expected disallow error, got: %s", out)
	}
}

func TestE2E_ValidateWithFormatEmail(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  ADMIN_EMAIL:
    type: string
    format: email
`
	env := "ADMIN_EMAIL=not-an-email\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "email") {
		t.Errorf("expected email format error, got: %s", out)
	}
}

func TestE2E_ValidateWithFormatURL(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  API_URL:
    type: string
    format: url
`
	env := "API_URL=not-a-url\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "URL") {
		t.Errorf("expected URL format error, got: %s", out)
	}
}

func TestE2E_ValidateWithFormatUUID(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  TRACE_ID:
    type: string
    format: uuid
`
	env := "TRACE_ID=not-a-uuid\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "UUID") {
		t.Errorf("expected UUID format error, got: %s", out)
	}
}

func TestE2E_ValidateWithFormatPort(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  PORT:
    type: string
    format: port
`
	env := "PORT=70000\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "port") {
		t.Errorf("expected port format error, got: %s", out)
	}
}

func TestE2E_ValidateWithFormatJSON(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  CONFIG:
    type: string
    format: json
`
	env := "CONFIG={invalid json}\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "JSON") {
		t.Errorf("expected JSON format error, got: %s", out)
	}
}

func TestE2E_ValidateWithFormatDuration(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  TIMEOUT:
    type: string
    format: duration
`
	env := "TIMEOUT=not-a-duration\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "duration") {
		t.Errorf("expected duration format error, got: %s", out)
	}
}

func TestE2E_ValidateWithFormatSemver(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  APP_VERSION:
    type: string
    format: semver
`
	env := "APP_VERSION=1.0\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "semantic version") {
		t.Errorf("expected semver format error, got: %s", out)
	}
}

func TestE2E_ValidateWithFormatHostname(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  HOST:
    type: string
    format: hostname
`
	env := "HOST=-invalid-hostname\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "hostname") {
		t.Errorf("expected hostname format error, got: %s", out)
	}
}

func TestE2E_ValidateWithFormatHex(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  COLOR:
    type: string
    format: hex
`
	env := "COLOR=GGGGGG\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "hexadecimal") {
		t.Errorf("expected hex format error, got: %s", out)
	}
}

func TestE2E_ValidateWithFormatCron(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  SCHEDULE:
    type: string
    format: cron
`
	env := "SCHEDULE=not-cron\n"

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "cron") {
		t.Errorf("expected cron format error, got: %s", out)
	}
}

func TestE2E_ValidateWithMinLengthMaxLength(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  CODE:
    type: string
    minLength: 3
    maxLength: 5
`
	env := "CODE=ab\n"

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

func TestE2E_ValidateWithIntegerMinMax(t *testing.T) {
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
	if !strings.Contains(out, "minimum") {
		t.Errorf("expected min error, got: %s", out)
	}
}

func TestE2E_ValidateWithFloatMinMax(t *testing.T) {
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
	if !strings.Contains(out, "maximum") {
		t.Errorf("expected max error, got: %s", out)
	}
}

func TestE2E_ValidateWithArrayMinMaxLength(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  ITEMS:
    type: array
    separator: ","
    minLength: 2
    maxLength: 4
`
	env := "ITEMS=a\n"

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

func TestE2E_ValidateWithArrayContains(t *testing.T) {
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

func TestE2E_ValidateWithCustomMessage(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	schema := `
version: "1.0"
env:
  API_KEY:
    type: string
    required: true
    message: "API_KEY is mandatory for all environments"
`
	env := ""

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "API_KEY is mandatory for all environments") {
		t.Errorf("expected custom message, got: %s", out)
	}
}

func TestE2E_ValidateWithAllowEmptyFalse(t *testing.T) {
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

func TestE2E_ValidateStrictModeUnknownKeys(t *testing.T) {
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
	if !strings.Contains(out, "UNKNOWN") {
		t.Errorf("expected UNKNOWN warning, got: %s", out)
	}
}

func TestE2E_ValidateMultipleErrorsCollected(t *testing.T) {
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
  D:
    type: float
    required: true
`
	env := `A=
B=not-an-int
C=not-a-bool
D=not-a-float
`

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(schema), 0644)
	os.WriteFile(envPath, []byte(env), 0644)

	out, code := runEnvGuard(t, bin, "validate", "--schema", schemaPath, "--env", envPath)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d. output: %s", code, out)
	}

	// All errors should be collected
	for _, key := range []string{"A", "B", "C", "D"} {
		if !strings.Contains(out, key) {
			t.Errorf("expected %s error in output, got: %s", key, out)
		}
	}
}

func TestE2E_InitCommandInTempDir(t *testing.T) {
	bin := buildEnvGuard(t)
	tmpDir := t.TempDir()

	// Change to tmpDir for init
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	out, code := runEnvGuard(t, bin, "init")
	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %s", code, out)
	}
	if !strings.Contains(out, "Created envguard.yaml") {
		t.Errorf("expected creation message, got: %s", out)
	}

	// Second init should fail
	out, code = runEnvGuard(t, bin, "init")
	if code != 1 {
		t.Errorf("expected exit code 1 for duplicate init, got %d. output: %s", code, out)
	}
}

func TestE2E_PreCommitHookStrict(t *testing.T) {
	content, err := os.ReadFile("../.pre-commit-hooks.yaml")
	if err != nil {
		t.Fatalf(".pre-commit-hooks.yaml should exist: %v", err)
	}
	if !strings.Contains(string(content), "--strict") {
		t.Errorf("pre-commit hook should use --strict flag")
	}
}
