package schema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseWithExtends(t *testing.T) {
	tmpDir := t.TempDir()

	base := `
version: "1.0"
env:
  DATABASE_URL:
    type: string
    required: true
  PORT:
    type: integer
    default: 3000
`
	extended := `
version: "1.0"
extends: ./base.yaml
env:
  API_KEY:
    type: string
    required: true
  PORT:
    type: integer
    default: 8080
`

	basePath := filepath.Join(tmpDir, "base.yaml")
	extendedPath := filepath.Join(tmpDir, "extended.yaml")
	os.WriteFile(basePath, []byte(base), 0644)
	os.WriteFile(extendedPath, []byte(extended), 0644)

	s, err := Parse(extendedPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := s.Env["DATABASE_URL"]; !ok {
		t.Error("DATABASE_URL should be inherited from base")
	}
	if _, ok := s.Env["API_KEY"]; !ok {
		t.Error("API_KEY should be defined in extended")
	}
	if s.Env["PORT"].Default != 8080 {
		t.Errorf("PORT default should be overridden to 8080, got %v", s.Env["PORT"].Default)
	}
}

func TestParseCircularExtends(t *testing.T) {
	tmpDir := t.TempDir()

	a := `
version: "1.0"
extends: ./b.yaml
env:
  A:
    type: string
`
	b := `
version: "1.0"
extends: ./a.yaml
env:
  B:
    type: string
`

	aPath := filepath.Join(tmpDir, "a.yaml")
	bPath := filepath.Join(tmpDir, "b.yaml")
	os.WriteFile(aPath, []byte(a), 0644)
	os.WriteFile(bPath, []byte(b), 0644)

	_, err := Parse(aPath)
	if err == nil {
		t.Fatal("expected error for circular extends")
	}
	if !strings.Contains(err.Error(), "circular") {
		t.Errorf("expected circular error, got: %v", err)
	}
}

func TestParseMissingExtends(t *testing.T) {
	tmpDir := t.TempDir()

	extended := `
version: "1.0"
extends: ./nonexistent.yaml
env:
  A:
    type: string
`

	extendedPath := filepath.Join(tmpDir, "extended.yaml")
	os.WriteFile(extendedPath, []byte(extended), 0644)

	_, err := Parse(extendedPath)
	if err == nil {
		t.Fatal("expected error for missing base schema")
	}
}
