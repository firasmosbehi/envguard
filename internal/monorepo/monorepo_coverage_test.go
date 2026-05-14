package monorepo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverNoEnvFiles(t *testing.T) {
	tmpDir := t.TempDir()

	projects, err := Discover(tmpDir, true)
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}
}

func TestDiscoverNestedDirs(t *testing.T) {
	tmpDir := t.TempDir()

	// Root .env
	os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("ROOT=1\n"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "envguard.yaml"), []byte("version: \"1.0\"\nenv:\n  ROOT:\n    type: string\n"), 0644)

	// Nested level 1
	sub1 := filepath.Join(tmpDir, "services", "api")
	os.MkdirAll(sub1, 0755)
	os.WriteFile(filepath.Join(sub1, ".env"), []byte("API=1\n"), 0644)
	os.WriteFile(filepath.Join(sub1, "envguard.yaml"), []byte("version: \"1.0\"\nenv:\n  API:\n    type: string\n"), 0644)

	// Nested level 2
	sub2 := filepath.Join(tmpDir, "services", "api", "v2")
	os.MkdirAll(sub2, 0755)
	os.WriteFile(filepath.Join(sub2, ".env"), []byte("V2=1\n"), 0644)
	os.WriteFile(filepath.Join(sub2, "envguard.yaml"), []byte("version: \"1.0\"\nenv:\n  V2:\n    type: string\n"), 0644)

	projects, err := Discover(tmpDir, true)
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}
	if len(projects) != 3 {
		t.Errorf("expected 3 projects, got %d", len(projects))
	}
}

func TestDiscoverSkipsIgnoredDirs(t *testing.T) {
	tmpDir := t.TempDir()

	// Root .env
	os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("ROOT=1\n"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "envguard.yaml"), []byte("version: \"1.0\"\n"), 0644)

	// node_modules should be skipped
	nodeModules := filepath.Join(tmpDir, "node_modules", "some-pkg")
	os.MkdirAll(nodeModules, 0755)
	os.WriteFile(filepath.Join(nodeModules, ".env"), []byte("NODE=1\n"), 0644)

	projects, err := Discover(tmpDir, true)
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}
	if len(projects) != 1 {
		t.Errorf("expected 1 project (node_modules skipped), got %d", len(projects))
	}
}

func TestFindSchemaForDirNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "a", "b", "c")
	os.MkdirAll(subDir, 0755)

	found := findSchemaForDir(subDir)
	if found != "" {
		t.Errorf("expected empty string, got %q", found)
	}
}

func TestValidateAllResultUsage(t *testing.T) {
	// Test that ValidateAllResult struct is usable
	result := ValidateAllResult{
		Projects: []ProjectResult{
			{Project: EnvProject{Dir: "/app1"}, Valid: true},
			{Project: EnvProject{Dir: "/app2"}, Valid: false, Errors: []string{"err1"}},
		},
		Valid: false,
	}

	if len(result.Projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(result.Projects))
	}
	if result.Valid {
		t.Error("expected Valid to be false")
	}
}

func TestFormatResultsAllValid(t *testing.T) {
	results := []ProjectResult{
		{Project: EnvProject{Dir: "/proj1"}, Valid: true},
		{Project: EnvProject{Dir: "/proj2"}, Valid: true},
	}
	out := FormatResults(results)
	if !contains(out, "2 passed, 0 failed") {
		t.Errorf("expected all-pass summary, got: %s", out)
	}
}

func TestFormatResultsAllInvalid(t *testing.T) {
	results := []ProjectResult{
		{Project: EnvProject{Dir: "/proj1"}, Valid: false, Errors: []string{"missing X"}},
		{Project: EnvProject{Dir: "/proj2"}, Valid: false, Errors: []string{"missing Y", "invalid Z"}},
	}
	out := FormatResults(results)
	if !contains(out, "0 passed, 2 failed") {
		t.Errorf("expected all-fail summary, got: %s", out)
	}
	if !contains(out, "missing X") {
		t.Error("expected error message in output")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
