package audit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGoExtractor(t *testing.T) {
	extractor := &goExtractor{}
	content := `
package main

import "os"

func main() {
	_ = os.Getenv("API_KEY")
	_ = os.Getenv("DB_HOST")
	_, _ = os.LookupEnv("CACHE_URL")
	os.Setenv("DEBUG", "true")
	os.Unsetenv("TEMP_VAR")
}
`
	refs := extractor.Extract("main.go", content)
	if len(refs) == 0 {
		t.Fatal("expected to find env refs")
	}

	expectedVars := map[string]bool{
		"API_KEY":   false,
		"DB_HOST":   false,
		"CACHE_URL": false,
		"DEBUG":     false,
		"TEMP_VAR":  false,
	}
	for _, ref := range refs {
		if _, ok := expectedVars[ref.Var]; ok {
			expectedVars[ref.Var] = true
		}
	}
	for v, found := range expectedVars {
		if !found {
			t.Errorf("expected to find %s", v)
		}
	}
}

func TestNodeExtractor(t *testing.T) {
	extractor := &nodeExtractor{}
	content := `
const apiKey = process.env.API_KEY;
const db = process.env["DATABASE_URL"];
const viteVar = import.meta.env.VITE_API_URL;
const denoVal = Deno.env.get("DENO_DEPLOY_TOKEN");
`
	refs := extractor.Extract("app.js", content)

	expected := []string{"API_KEY", "DATABASE_URL", "VITE_API_URL", "DENO_DEPLOY_TOKEN"}
	found := make(map[string]bool)
	for _, ref := range refs {
		found[ref.Var] = true
	}
	for _, v := range expected {
		if !found[v] {
			t.Errorf("expected to find %s", v)
		}
	}
}

func TestPythonExtractor(t *testing.T) {
	extractor := &pythonExtractor{}
	content := `
import os

api_key = os.environ['API_KEY']
db_url = os.environ.get("DATABASE_URL")
port = os.getenv('PORT')
os.putenv("DEBUG", "1")
`
	refs := extractor.Extract("app.py", content)

	expected := []string{"API_KEY", "DATABASE_URL", "PORT", "DEBUG"}
	found := make(map[string]bool)
	for _, ref := range refs {
		found[ref.Var] = true
	}
	for _, v := range expected {
		if !found[v] {
			t.Errorf("expected to find %s", v)
		}
	}
}

func TestRustExtractor(t *testing.T) {
	extractor := &rustExtractor{}
	content := `
let api_key = std::env::var("API_KEY").unwrap();
let path = std::env::var_os("PATH").unwrap();
std::env::set_var("DEBUG", "1");
`
	refs := extractor.Extract("main.rs", content)

	expected := []string{"API_KEY", "PATH", "DEBUG"}
	found := make(map[string]bool)
	for _, ref := range refs {
		found[ref.Var] = true
	}
	for _, v := range expected {
		if !found[v] {
			t.Errorf("expected to find %s", v)
		}
	}
}

func TestRubyExtractor(t *testing.T) {
	extractor := &rubyExtractor{}
	content := `
api_key = ENV['API_KEY']
db_url = ENV.fetch("DATABASE_URL")
port = ENV.fetch("PORT", "3000")
`
	refs := extractor.Extract("app.rb", content)

	expected := []string{"API_KEY", "DATABASE_URL", "PORT"}
	found := make(map[string]bool)
	for _, ref := range refs {
		found[ref.Var] = true
	}
	for _, v := range expected {
		if !found[v] {
			t.Errorf("expected to find %s", v)
		}
	}
}

func TestJavaExtractor(t *testing.T) {
	extractor := &javaExtractor{}
	content := `
String apiKey = System.getenv("API_KEY");
String port = System.getProperty("PORT");
`
	refs := extractor.Extract("App.java", content)

	expected := []string{"API_KEY", "PORT"}
	found := make(map[string]bool)
	for _, ref := range refs {
		found[ref.Var] = true
	}
	for _, v := range expected {
		if !found[v] {
			t.Errorf("expected to find %s", v)
		}
	}
}

func TestRun_MissingVar(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a Go source file referencing an env var
	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "main.go"), []byte(`
package main
import "os"
func main() {
	_ = os.Getenv("MISSING_VAR")
}
`), 0644)

	// Create .env without MISSING_VAR
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte("EXISTING=1\n"), 0644)

	result, err := Run(Options{
		SrcDir:     srcDir,
		EnvPath:    envPath,
		SchemaPath: "",
		Exclude:    nil,
		IgnoreVars: []string{},
	})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	found := false
	for _, f := range result.Findings {
		if f.Type == MissingVar && f.Var == "MISSING_VAR" {
			found = true
			if f.File != "main.go" {
				t.Errorf("expected file main.go, got %q", f.File)
			}
			if f.Line != 5 {
				t.Errorf("expected line 5, got %d", f.Line)
			}
		}
	}
	if !found {
		t.Errorf("expected MissingVar finding for MISSING_VAR, got: %+v", result.Findings)
	}
}

func TestRun_UnusedVar(t *testing.T) {
	tmpDir := t.TempDir()

	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "main.go"), []byte(`
package main
func main() {}
`), 0644)

	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte("UNUSED_VAR=1\n"), 0644)

	result, err := Run(Options{
		SrcDir:     srcDir,
		EnvPath:    envPath,
		SchemaPath: "",
		Exclude:    nil,
		IgnoreVars: []string{},
	})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	found := false
	for _, f := range result.Findings {
		if f.Type == UnusedVar && f.Var == "UNUSED_VAR" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected UnusedVar finding for UNUSED_VAR, got: %+v", result.Findings)
	}
}

func TestRun_UndocumentedVar(t *testing.T) {
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
		Exclude:    nil,
		IgnoreVars: []string{},
	})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

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

func TestRun_MissingRequired(t *testing.T) {
	tmpDir := t.TempDir()

	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "main.go"), []byte(`
package main
func main() {}
`), 0644)

	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte("EXISTING=1\n"), 0644)

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte(`
version: "1.0"
env:
  REQUIRED_VAR:
    type: string
    required: true
`), 0644)

	result, err := Run(Options{
		SrcDir:     srcDir,
		EnvPath:    envPath,
		SchemaPath: schemaPath,
		Exclude:    nil,
		IgnoreVars: []string{},
	})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	found := false
	for _, f := range result.Findings {
		if f.Type == MissingRequired && f.Var == "REQUIRED_VAR" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected MissingRequired finding for REQUIRED_VAR, got: %+v", result.Findings)
	}
}

func TestRun_IgnoreVars(t *testing.T) {
	tmpDir := t.TempDir()

	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "main.go"), []byte(`
package main
import "os"
func main() {
	_ = os.Getenv("HOME")
	_ = os.Getenv("MY_VAR")
}
`), 0644)

	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte("MY_VAR=1\n"), 0644)

	result, err := Run(Options{
		SrcDir:     srcDir,
		EnvPath:    envPath,
		SchemaPath: "",
		Exclude:    nil,
		IgnoreVars: []string{"HOME"},
	})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	for _, f := range result.Findings {
		if f.Var == "HOME" {
			t.Errorf("HOME should be ignored, but found: %+v", f)
		}
	}
}

func TestRun_Exclude(t *testing.T) {
	tmpDir := t.TempDir()

	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)
	os.MkdirAll(filepath.Join(srcDir, "vendor"), 0755)
	os.WriteFile(filepath.Join(srcDir, "main.go"), []byte(`
package main
import "os"
func main() {
	_ = os.Getenv("MAIN_VAR")
}
`), 0644)
	os.WriteFile(filepath.Join(srcDir, "vendor", "lib.go"), []byte(`
package vendor
import "os"
func init() {
	_ = os.Getenv("VENDOR_VAR")
}
`), 0644)

	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte("MAIN_VAR=1\nVENDOR_VAR=1\n"), 0644)

	result, err := Run(Options{
		SrcDir:     srcDir,
		EnvPath:    envPath,
		SchemaPath: "",
		Exclude:    []string{"vendor/"},
		IgnoreVars: []string{"VENDOR_VAR"},
	})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	for _, f := range result.Findings {
		if f.Var == "VENDOR_VAR" {
			t.Errorf("VENDOR_VAR should be ignored, but found: %+v", f)
		}
	}
}

func TestDeduplicateRefs(t *testing.T) {
	refs := []EnvRef{
		{Var: "A", File: "f.go", Line: 1},
		{Var: "A", File: "f.go", Line: 1},
		{Var: "A", File: "f.go", Line: 2},
		{Var: "B", File: "f.go", Line: 1},
	}
	deduped := deduplicateRefs(refs)
	if len(deduped) != 3 {
		t.Errorf("expected 3 refs after dedup, got %d", len(deduped))
	}
}

func TestResultCountByType(t *testing.T) {
	r := NewResult()
	r.AddFinding(Finding{Type: MissingVar})
	r.AddFinding(Finding{Type: MissingVar})
	r.AddFinding(Finding{Type: UnusedVar})

	if r.CountByType(MissingVar) != 2 {
		t.Errorf("expected 2 MissingVar, got %d", r.CountByType(MissingVar))
	}
	if r.CountByType(UnusedVar) != 1 {
		t.Errorf("expected 1 UnusedVar, got %d", r.CountByType(UnusedVar))
	}
	if r.CountByType(UndocumentedVar) != 0 {
		t.Errorf("expected 0 UndocumentedVar, got %d", r.CountByType(UndocumentedVar))
	}
}
