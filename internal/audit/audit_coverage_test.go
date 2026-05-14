package audit

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestKnownRuntimeVars(t *testing.T) {
	vars := KnownRuntimeVars()
	if len(vars) == 0 {
		t.Error("expected non-empty list of known runtime vars")
	}

	// Check some expected vars
	expected := map[string]bool{
		"PATH":     false,
		"HOME":     false,
		"NODE_ENV": false,
		"CI":       false,
	}

	for _, v := range vars {
		if _, ok := expected[v]; ok {
			expected[v] = true
		}
	}

	for v, found := range expected {
		if !found {
			t.Errorf("expected %s in known runtime vars", v)
		}
	}
}

func TestHasFindings(t *testing.T) {
	r := NewResult()
	if r.HasFindings() {
		t.Error("expected HasFindings to be false for empty result")
	}

	r.AddFinding(Finding{Type: MissingVar})
	if !r.HasFindings() {
		t.Error("expected HasFindings to be true after adding finding")
	}
}

func TestRunEmptySrcDir(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)

	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte("VAR=1\n"), 0644)

	result, err := Run(Options{
		SrcDir:     srcDir,
		EnvPath:    envPath,
		SchemaPath: "",
	})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// VAR is in .env but not in code -> unused
	found := false
	for _, f := range result.Findings {
		if f.Type == UnusedVar && f.Var == "VAR" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected UnusedVar finding for VAR, got: %+v", result.Findings)
	}
}

func TestRunWithSchemaAndCodeVar(t *testing.T) {
	tmpDir := t.TempDir()

	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "main.go"), []byte(`
package main
import "os"
func main() {
	_ = os.Getenv("CODE_VAR")
}
`), 0644)

	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte("CODE_VAR=1\n"), 0644)

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte(`
version: "1.0"
env:
  OTHER_VAR:
    type: string
`), 0644)

	result, err := Run(Options{
		SrcDir:     srcDir,
		EnvPath:    envPath,
		SchemaPath: schemaPath,
	})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// CODE_VAR is in code and .env but not in schema -> undocumented
	found := false
	for _, f := range result.Findings {
		if f.Type == UndocumentedVar && f.Var == "CODE_VAR" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected UndocumentedVar finding for CODE_VAR, got: %+v", result.Findings)
	}
}

func TestWalkFilesInvalidExcludePattern(t *testing.T) {
	_, err := walkFiles(".", []string{"[invalid("})
	if err == nil {
		t.Error("expected error for invalid exclude pattern")
	}
}

func TestWalkFilesUnreadableDir(t *testing.T) {
	// Create a directory that we can't read
	tmpDir := t.TempDir()
	unreadableDir := filepath.Join(tmpDir, "unreadable")
	os.MkdirAll(unreadableDir, 0755)
	os.WriteFile(filepath.Join(unreadableDir, "test.go"), []byte("package main\n"), 0644)
	os.Chmod(unreadableDir, 0000)
	defer os.Chmod(unreadableDir, 0755) // restore for cleanup

	_, err := walkFiles(tmpDir, nil)
	if err != nil {
		t.Fatalf("walkFiles should skip unreadable dirs, got: %v", err)
	}
}

func TestRunStrictMode(t *testing.T) {
	tmpDir := t.TempDir()

	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "main.go"), []byte(`
package main
import "os"
func main() {
	_ = os.Getenv("HOME")
}
`), 0644)

	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte("HOME=/user\n"), 0644)

	result, err := Run(Options{
		SrcDir:     srcDir,
		EnvPath:    envPath,
		SchemaPath: "",
		Strict:     true,
		IgnoreVars: []string{"HOME"},
	})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// HOME is ignored, so no findings should mention it
	for _, f := range result.Findings {
		if f.Var == "HOME" {
			t.Errorf("HOME should be ignored, but found: %+v", f)
		}
	}
}

func TestExtractRegexNoCaptureGroup(t *testing.T) {
	// Test extractRegex with regex that has no capture groups
	re := regexp.MustCompile(`os\.Getenv\("[^"]+"\)`)
	refs := extractRegex("main.go", `_ = os.Getenv("TEST")`, re)
	if len(refs) != 0 {
		t.Errorf("expected 0 refs without capture group, got %d", len(refs))
	}
}
