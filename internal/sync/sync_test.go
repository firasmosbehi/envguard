package sync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_SyncExample(t *testing.T) {
	tmpDir := t.TempDir()

	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte("API_KEY=secret123\nDB_HOST=localhost\n"), 0644)

	examplePath := filepath.Join(tmpDir, ".env.example")
	os.WriteFile(examplePath, []byte("API_KEY=\n"), 0644)

	result, err := Run(Options{
		EnvPath:     envPath,
		ExamplePath: examplePath,
		SchemaPath:  "",
		Check:       false,
	})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if !result.WouldWrite {
		t.Error("expected WouldWrite to be true")
	}

	// Check that DB_HOST was added to example
	content, _ := os.ReadFile(examplePath)
	if !strings.Contains(string(content), "DB_HOST=") {
		t.Errorf("expected .env.example to contain DB_HOST, got:\n%s", string(content))
	}
}

func TestRun_CheckMode(t *testing.T) {
	tmpDir := t.TempDir()

	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte("API_KEY=secret123\nDB_HOST=localhost\n"), 0644)

	examplePath := filepath.Join(tmpDir, ".env.example")
	os.WriteFile(examplePath, []byte("API_KEY=\n"), 0644)

	result, err := Run(Options{
		EnvPath:     envPath,
		ExamplePath: examplePath,
		SchemaPath:  "",
		Check:       true,
	})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if !result.WouldWrite {
		t.Error("expected WouldWrite to be true in check mode when drift exists")
	}

	// Ensure example file was NOT modified
	content, _ := os.ReadFile(examplePath)
	if strings.Contains(string(content), "DB_HOST") {
		t.Error("check mode should not modify .env.example")
	}
}

func TestRun_NoDrift(t *testing.T) {
	tmpDir := t.TempDir()

	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte("API_KEY=secret123\n"), 0644)

	examplePath := filepath.Join(tmpDir, ".env.example")
	os.WriteFile(examplePath, []byte("API_KEY=secret123\n"), 0644)

	result, err := Run(Options{
		EnvPath:     envPath,
		ExamplePath: examplePath,
		SchemaPath:  "",
		Check:       true,
	})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result.WouldWrite {
		t.Error("expected WouldWrite to be false when no drift")
	}
	if len(result.Diffs) != 0 {
		t.Errorf("expected 0 diffs, got %d", len(result.Diffs))
	}
}

func TestRun_WithSchema(t *testing.T) {
	tmpDir := t.TempDir()

	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte("API_KEY=secret123\n"), 0644)

	examplePath := filepath.Join(tmpDir, ".env.example")
	os.WriteFile(examplePath, []byte("API_KEY=\n"), 0644)

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte(`
version: "1.0"
env:
  SCHEMA_VAR:
    type: string
    required: true
  API_KEY:
    type: string
    sensitive: true
`), 0644)

	result, err := Run(Options{
		EnvPath:     envPath,
		ExamplePath: examplePath,
		SchemaPath:  schemaPath,
		Check:       false,
	})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	_ = result

	// Check that SCHEMA_VAR was added
	content, _ := os.ReadFile(examplePath)
	if !strings.Contains(string(content), "SCHEMA_VAR=") {
		t.Errorf("expected .env.example to contain SCHEMA_VAR, got:\n%s", string(content))
	}
	// Check that API_KEY was masked
	if strings.Contains(string(content), "secret123") {
		t.Error("sensitive API_KEY should be masked in .env.example")
	}
	if !strings.Contains(string(content), "API_KEY=***") {
		t.Errorf("expected API_KEY to be masked, got:\n%s", string(content))
	}
}

func TestRun_AddMissing(t *testing.T) {
	tmpDir := t.TempDir()

	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte("API_KEY=secret123\n"), 0644)

	examplePath := filepath.Join(tmpDir, ".env.example")
	os.WriteFile(examplePath, []byte("API_KEY=\nEXTRA_VAR=hello\n"), 0644)

	result, err := Run(Options{
		EnvPath:     envPath,
		ExamplePath: examplePath,
		SchemaPath:  "",
		Check:       false,
		AddMissing:  true,
	})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	_ = result

	// Check that EXTRA_VAR was added to .env
	envContent, _ := os.ReadFile(envPath)
	if !strings.Contains(string(envContent), "EXTRA_VAR=") {
		t.Errorf("expected .env to contain EXTRA_VAR after --add-missing, got:\n%s", string(envContent))
	}
}

func TestMaskValue(t *testing.T) {
	if maskValue("secret") != "***" {
		t.Errorf("expected ***, got %q", maskValue("secret"))
	}
	if maskValue("") != "" {
		t.Errorf("expected empty string, got %q", maskValue(""))
	}
}

func TestSuggestPlaceholder(t *testing.T) {
	tests := []struct {
		key  string
		val  string
		want string
	}{
		{"DATABASE_URL", "postgres://localhost", "postgresql://user:password@localhost:5432/dbname"},
		{"REDIS_URL", "redis://localhost", "redis://localhost:6379"},
		{"API_KEY", "abc123", "your-api-key-here"},
		{"PORT", "3000", "3000"},
		{"URL", "https://example.com", "https://example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := suggestPlaceholder(tt.key, tt.val, nil)
			if got != tt.want {
				t.Errorf("suggestPlaceholder(%q, %q) = %q, want %q", tt.key, tt.val, got, tt.want)
			}
		})
	}
}

func TestParseEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".env")
	content := "# Comment\nKEY1=value1\n\nKEY2='quoted'\nKEY3=\"double\"\n"
	os.WriteFile(path, []byte(content), 0644)

	entries, lookup, err := parseEnvFile(path)
	if err != nil {
		t.Fatalf("parseEnvFile failed: %v", err)
	}

	if len(entries) != 5 { // comment + 3 keys + 1 blank
		t.Errorf("expected 5 entries, got %d", len(entries))
	}

	if lookup["KEY1"].Value != "value1" {
		t.Errorf("expected KEY1=value1, got %q", lookup["KEY1"].Value)
	}
	if lookup["KEY2"].Value != "quoted" {
		t.Errorf("expected KEY2=quoted, got %q", lookup["KEY2"].Value)
	}
	if lookup["KEY3"].Value != "double" {
		t.Errorf("expected KEY3=double, got %q", lookup["KEY3"].Value)
	}
}
