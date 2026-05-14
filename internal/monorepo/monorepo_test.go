package monorepo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscover(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("ROOT=1\n"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "envguard.yaml"), []byte("version: \"1.0\"\nenv:\n  ROOT:\n    type: string\n"), 0644)

	subDir := filepath.Join(tmpDir, "api")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, ".env"), []byte("API=1\n"), 0644)
	os.WriteFile(filepath.Join(subDir, "envguard.yaml"), []byte("version: \"1.0\"\nenv:\n  API:\n    type: string\n"), 0644)

	projects, err := Discover(tmpDir, false)
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}
	if len(projects) != 1 {
		t.Errorf("expected 1 project non-recursive, got %d", len(projects))
	}

	projects, err = Discover(tmpDir, true)
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}
	if len(projects) != 2 {
		t.Errorf("expected 2 projects recursive, got %d", len(projects))
	}

	for _, p := range projects {
		if p.SchemaPath == "" {
			t.Errorf("expected schema path for project in %s", p.Dir)
		}
	}
}

func TestFindSchemaForDir(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "a", "b")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(tmpDir, "envguard.yaml"), []byte("version: \"1.0\"\n"), 0644)

	found := findSchemaForDir(subDir)
	expected := filepath.Join(tmpDir, "envguard.yaml")
	if found != expected {
		t.Errorf("expected %q, got %q", expected, found)
	}
}

func TestFormatResults(t *testing.T) {
	results := []ProjectResult{
		{Project: EnvProject{Dir: "/proj1"}, Valid: true},
		{Project: EnvProject{Dir: "/proj2"}, Valid: false, Errors: []string{"missing API_KEY"}},
	}
	out := FormatResults(results)
	if !strings.Contains(out, "1 passed, 1 failed") {
		t.Errorf("expected summary, got: %s", out)
	}
}
