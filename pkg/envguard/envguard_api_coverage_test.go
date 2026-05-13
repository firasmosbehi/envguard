package envguard

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateFileStrictMode(t *testing.T) {
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

	result, err := ValidateFile(schemaPath, envPath, true, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid (warnings only), got errors: %v", result.Errors)
	}
	if len(result.Warnings) != 1 {
		t.Errorf("expected 1 warning, got %d", len(result.Warnings))
	}
}

func TestValidateFileWithEnvName(t *testing.T) {
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

	result, err := ValidateFile(schemaPath, envPath, false, "production")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Valid {
		t.Error("expected invalid in production")
	}

	result, err = ValidateFile(schemaPath, envPath, false, "development")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid in development, got errors: %v", result.Errors)
	}
}

func TestValidateFileMissingSchema(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte("FOO=bar\n"), 0644)

	_, err := ValidateFile(filepath.Join(tmpDir, "nonexistent.yaml"), envPath, false, "")
	if err == nil {
		t.Fatal("expected error for missing schema")
	}
}

func TestValidateFileMissingEnv(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte(`
version: "1.0"
env:
  FOO:
    type: string
`), 0644)

	_, err := ValidateFile(schemaPath, filepath.Join(tmpDir, "nonexistent.env"), false, "")
	if err == nil {
		t.Fatal("expected error for missing env file")
	}
}

func TestValidateFileInvalidSchemaYAML(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(`not: [ valid yaml :::`), 0644)
	os.WriteFile(envPath, []byte("FOO=bar\n"), 0644)

	_, err := ValidateFile(schemaPath, envPath, false, "")
	if err == nil {
		t.Fatal("expected error for invalid schema YAML")
	}
}

func TestValidateFileInvalidSchemaStructure(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(`
version: "1.0"
env:
  FOO:
    type: invalidtype
`), 0644)
	os.WriteFile(envPath, []byte("FOO=bar\n"), 0644)

	_, err := ValidateFile(schemaPath, envPath, false, "")
	if err == nil {
		t.Fatal("expected error for invalid schema structure")
	}
}

func TestParseSchemaMissingFile(t *testing.T) {
	_, err := ParseSchema("/nonexistent/schema.yaml")
	if err == nil {
		t.Fatal("expected error for missing schema file")
	}
}

func TestParseSchemaInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte(`not: [ valid yaml :::`), 0644)

	_, err := ParseSchema(schemaPath)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestParseEnvMissingFile(t *testing.T) {
	_, err := ParseEnv("/nonexistent/.env")
	if err == nil {
		t.Fatal("expected error for missing env file")
	}
}

func TestValidateDirect(t *testing.T) {
	schema, err := ParseSchema("../../examples/envguard.yaml")
	if err != nil {
		t.Fatalf("failed to parse example schema: %v", err)
	}

	envVars, err := ParseEnv("../../examples/.env")
	if err != nil {
		t.Fatalf("failed to parse example env: %v", err)
	}

	result := Validate(schema, envVars, false, "")
	if !result.Valid {
		t.Errorf("expected example files to validate, got errors: %v", result.Errors)
	}
}

func TestValidateDirectInvalid(t *testing.T) {
	schema, err := ParseSchema("../../examples/envguard.yaml")
	if err != nil {
		t.Fatalf("failed to parse example schema: %v", err)
	}

	envVars, err := ParseEnv("../../examples/.env.invalid")
	if err != nil {
		t.Fatalf("failed to parse invalid example env: %v", err)
	}

	result := Validate(schema, envVars, false, "")
	if result.Valid {
		t.Error("expected invalid example to fail validation")
	}
}
