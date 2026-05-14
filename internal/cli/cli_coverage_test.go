package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/envguard/envguard/internal/audit"
	"github.com/envguard/envguard/internal/schema"
	"github.com/envguard/envguard/internal/secrets"
	envguardsync "github.com/envguard/envguard/internal/sync"
)

// === audit.go tests ===

func TestAuditCommandNoFindings(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")

	os.WriteFile(envPath, []byte(""), 0644)
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  FOO:
    type: string
`), 0644)

	opts := &auditOptions{
		srcDir:     tmpDir,
		envPath:    envPath,
		schemaPath: schemaPath,
		format:     "text",
	}
	var stdout, stderr bytes.Buffer
	err := runAudit(&stdout, &stderr, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "No issues found") {
		t.Errorf("expected no issues, got: %s", stdout.String())
	}
}

func TestAuditCommandWithFindings(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)

	goFile := filepath.Join(srcDir, "main.go")
	os.WriteFile(goFile, []byte(`package main
import "os"
func main() {
    _ = os.Getenv("MISSING_VAR")
}
`), 0644)

	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte("OTHER=val\n"), 0644)

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  REQUIRED_VAR:
    type: string
    required: true
`), 0644)

	opts := &auditOptions{
		srcDir:     srcDir,
		envPath:    envPath,
		schemaPath: schemaPath,
		format:     "text",
	}
	var stdout, stderr bytes.Buffer
	err := runAudit(&stdout, &stderr, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "MISSING_VAR") {
		t.Errorf("expected MISSING_VAR finding, got: %s", out)
	}
	if !strings.Contains(out, "REQUIRED_VAR") {
		t.Errorf("expected REQUIRED_VAR finding, got: %s", out)
	}
	if !strings.Contains(out, "OTHER") {
		t.Errorf("expected OTHER unused finding, got: %s", out)
	}
}

func TestAuditCommandJSON(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)

	goFile := filepath.Join(srcDir, "main.go")
	os.WriteFile(goFile, []byte(`package main
import "os"
func main() {
    _ = os.Getenv("MISSING_VAR")
}
`), 0644)

	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte(""), 0644)

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  FOO:
    type: string
`), 0644)

	opts := &auditOptions{
		srcDir:     srcDir,
		envPath:    envPath,
		schemaPath: schemaPath,
		format:     "json",
	}
	var stdout, stderr bytes.Buffer
	err := runAudit(&stdout, &stderr, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), `"var"`) {
		t.Errorf("expected JSON output, got: %s", stdout.String())
	}
}

func TestAuditCommandSARIF(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)

	goFile := filepath.Join(srcDir, "main.go")
	os.WriteFile(goFile, []byte(`package main
import "os"
func main() {
    _ = os.Getenv("MISSING_VAR")
}
`), 0644)

	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte(""), 0644)

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  FOO:
    type: string
`), 0644)

	opts := &auditOptions{
		srcDir:     srcDir,
		envPath:    envPath,
		schemaPath: schemaPath,
		format:     "sarif",
	}
	var stdout, stderr bytes.Buffer
	err := runAudit(&stdout, &stderr, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), `"$schema"`) {
		t.Errorf("expected SARIF output, got: %s", stdout.String())
	}
}

func TestAuditCommandUnknownFormat(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)

	goFile := filepath.Join(srcDir, "main.go")
	os.WriteFile(goFile, []byte(`package main
import "os"
func main() {
    _ = os.Getenv("MISSING_VAR")
}
`), 0644)

	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte(""), 0644)

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  FOO:
    type: string
`), 0644)

	opts := &auditOptions{
		srcDir:     srcDir,
		envPath:    envPath,
		schemaPath: schemaPath,
		format:     "xml",
	}
	var stdout, stderr bytes.Buffer
	err := runAudit(&stdout, &stderr, opts)
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
}

func TestAuditCommandStrict(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)

	goFile := filepath.Join(srcDir, "main.go")
	os.WriteFile(goFile, []byte(`package main
import "os"
func main() {
    _ = os.Getenv("MISSING_VAR")
}
`), 0644)

	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte(""), 0644)

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  REQUIRED_VAR:
    type: string
    required: true
`), 0644)

	opts := &auditOptions{
		srcDir:     srcDir,
		envPath:    envPath,
		schemaPath: schemaPath,
		format:     "text",
		strict:     true,
	}
	var stdout, stderr bytes.Buffer
	err := runAudit(&stdout, &stderr, opts)
	if err == nil {
		t.Fatal("expected error in strict mode")
	}
}

func TestAuditCommandError(t *testing.T) {
	tmpDir := t.TempDir()
	opts := &auditOptions{
		srcDir:  tmpDir,
		exclude: []string{"[invalid"},
		format:  "text",
	}
	var stdout, stderr bytes.Buffer
	err := runAudit(&stdout, &stderr, opts)
	if err == nil {
		t.Fatal("expected error for invalid exclude pattern")
	}
}

func TestPrintAuditText(t *testing.T) {
	result := &audit.Result{
		Findings: []audit.Finding{
			{Type: audit.MissingVar, Var: "V1", File: "main.go", Line: 10, Message: "missing"},
			{Type: audit.UnusedVar, Var: "V2", Message: "unused"},
			{Type: audit.UndocumentedVar, Var: "V3", File: "main.go", Line: 20, Message: "undoc"},
			{Type: audit.MissingRequired, Var: "V4", Message: "required"},
		},
	}
	var buf bytes.Buffer
	printAuditText(&buf, result)
	out := buf.String()
	if !strings.Contains(out, "Missing") {
		t.Errorf("expected Missing section, got: %s", out)
	}
	if !strings.Contains(out, "Unused") {
		t.Errorf("expected Unused section, got: %s", out)
	}
	if !strings.Contains(out, "Undocumented") {
		t.Errorf("expected Undocumented section, got: %s", out)
	}
	if !strings.Contains(out, "Missing Required") {
		t.Errorf("expected Missing Required section, got: %s", out)
	}
}

// === docs.go tests ===

func TestDocsCommandMarkdown(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  FOO:
    type: string
    description: "A foo"
`), 0644)

	opts := &docsOptions{
		schemaPath: schemaPath,
		format:     "markdown",
	}
	var stdout, stderr bytes.Buffer
	err := runDocs(&stdout, &stderr, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "FOO") {
		t.Errorf("expected FOO in docs, got: %s", stdout.String())
	}
}

func TestDocsCommandHTMLOutput(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  FOO:
    type: string
    description: "A foo"
`), 0644)

	opts := &docsOptions{
		schemaPath: schemaPath,
		format:     "html",
	}
	var stdout, stderr bytes.Buffer
	err := runDocs(&stdout, &stderr, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "<!DOCTYPE html>") {
		t.Errorf("expected HTML output, got: %s", stdout.String())
	}
}

func TestDocsCommandJSONOutput(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  FOO:
    type: string
    description: "A foo"
`), 0644)

	opts := &docsOptions{
		schemaPath: schemaPath,
		format:     "json",
	}
	var stdout, stderr bytes.Buffer
	err := runDocs(&stdout, &stderr, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), `"version"`) {
		t.Errorf("expected JSON output, got: %s", stdout.String())
	}
}

func TestDocsCommandOutputFile(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	outputPath := filepath.Join(tmpDir, "out.md")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  FOO:
    type: string
`), 0644)

	opts := &docsOptions{
		schemaPath: schemaPath,
		format:     "markdown",
		output:     outputPath,
	}
	var stdout, stderr bytes.Buffer
	err := runDocs(&stdout, &stderr, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "Generated markdown documentation") {
		t.Errorf("expected success message, got: %s", stdout.String())
	}
	content, _ := os.ReadFile(outputPath)
	if !strings.Contains(string(content), "FOO") {
		t.Errorf("expected FOO in output file")
	}
}

func TestDocsCommandMissingSchema(t *testing.T) {
	opts := &docsOptions{
		schemaPath: "/nonexistent/schema.yaml",
		format:     "markdown",
	}
	var stdout, stderr bytes.Buffer
	err := runDocs(&stdout, &stderr, opts)
	if err == nil {
		t.Fatal("expected error for missing schema")
	}
}

func TestDocsCommandGroupByPrefix(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  DB_HOST:
    type: string
  DB_PORT:
    type: integer
  API_KEY:
    type: string
`), 0644)

	opts := &docsOptions{
		schemaPath: schemaPath,
		format:     "markdown",
		groupBy:    "prefix",
	}
	var stdout, stderr bytes.Buffer
	err := runDocs(&stdout, &stderr, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "DB") {
		t.Errorf("expected DB group, got: %s", stdout.String())
	}
}

// === hook.go tests ===

func TestInstallHookCommand(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	if err := exec.Command("git", "init").Run(); err != nil {
		t.Skip("git not available")
	}

	opts := &struct {
		hookType string
		force    bool
		command  string
	}{
		hookType: "pre-commit",
	}
	var stdout, stderr bytes.Buffer
	err := runInstallHook(&stdout, &stderr, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "Installed pre-commit hook") {
		t.Errorf("expected install message, got: %s", stdout.String())
	}
}

func TestInstallHookCommandForce(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	if err := exec.Command("git", "init").Run(); err != nil {
		t.Skip("git not available")
	}

	opts := &struct {
		hookType string
		force    bool
		command  string
	}{
		hookType: "pre-commit",
	}
	var stdout, stderr bytes.Buffer
	// Install once
	runInstallHook(&stdout, &stderr, opts)
	stdout.Reset()
	stderr.Reset()
	// Install again with force
	opts.force = true
	err := runInstallHook(&stdout, &stderr, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "Installed pre-commit hook") {
		t.Errorf("expected install message, got: %s", stdout.String())
	}
}

func TestInstallHookCommandUnsupported(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	if err := exec.Command("git", "init").Run(); err != nil {
		t.Skip("git not available")
	}

	opts := &struct {
		hookType string
		force    bool
		command  string
	}{
		hookType: "invalid",
	}
	var stdout, stderr bytes.Buffer
	err := runInstallHook(&stdout, &stderr, opts)
	if err == nil {
		t.Fatal("expected error for unsupported hook type")
	}
}

func TestInstallHookCommandNoGit(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	opts := &struct {
		hookType string
		force    bool
		command  string
	}{
		hookType: "pre-commit",
	}
	var stdout, stderr bytes.Buffer
	err := runInstallHook(&stdout, &stderr, opts)
	if err == nil {
		t.Fatal("expected error when not in git repo")
	}
}

func TestUninstallHookCommand(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	if err := exec.Command("git", "init").Run(); err != nil {
		t.Skip("git not available")
	}

	// Install first
	opts := &struct {
		hookType string
		force    bool
		command  string
	}{
		hookType: "pre-commit",
	}
	var stdout, stderr bytes.Buffer
	runInstallHook(&stdout, &stderr, opts)

	// Uninstall via command
	cmd := newUninstallHookCmd()
	cmd.SetArgs([]string{})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "Uninstalled pre-commit hook") {
		t.Errorf("expected uninstall message, got: %s", buf.String())
	}
}

func TestUninstallHookCommandMissing(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	if err := exec.Command("git", "init").Run(); err != nil {
		t.Skip("git not available")
	}

	cmd := newUninstallHookCmd()
	cmd.SetArgs([]string{})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing hook")
	}
}

// === sync.go tests ===

func TestSyncCommandInSync(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	os.WriteFile(".env", []byte("FOO=bar\n"), 0644)
	os.WriteFile(".env.example", []byte("FOO=bar\n"), 0644)

	opts := &syncOptions{
		envPath:     ".env",
		examplePath: ".env.example",
		format:      "text",
	}
	var stdout, stderr bytes.Buffer
	err := runSync(&stdout, &stderr, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "in sync") {
		t.Errorf("expected in sync message, got: %s", stdout.String())
	}
}

func TestSyncCommandWithDiffs(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	os.WriteFile(".env", []byte("FOO=bar\nBAZ=qux\n"), 0644)
	os.WriteFile(".env.example", []byte("FOO=bar\n"), 0644)

	opts := &syncOptions{
		envPath:     ".env",
		examplePath: ".env.example",
		format:      "text",
	}
	var stdout, stderr bytes.Buffer
	err := runSync(&stdout, &stderr, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "Updated") {
		t.Errorf("expected updated message, got: %s", stdout.String())
	}
}

func TestSyncCommandJSON(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	os.WriteFile(".env", []byte("FOO=bar\n"), 0644)
	os.WriteFile(".env.example", []byte("BAZ=qux\n"), 0644)

	opts := &syncOptions{
		envPath:     ".env",
		examplePath: ".env.example",
		format:      "json",
	}
	var stdout, stderr bytes.Buffer
	err := runSync(&stdout, &stderr, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), `"diffs"`) {
		t.Errorf("expected JSON output, got: %s", stdout.String())
	}
}

func TestSyncCommandSARIF(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	os.WriteFile(".env", []byte("FOO=bar\n"), 0644)
	os.WriteFile(".env.example", []byte("BAZ=qux\n"), 0644)

	opts := &syncOptions{
		envPath:     ".env",
		examplePath: ".env.example",
		format:      "sarif",
	}
	var stdout, stderr bytes.Buffer
	err := runSync(&stdout, &stderr, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), `"$schema"`) {
		t.Errorf("expected SARIF output, got: %s", stdout.String())
	}
}

func TestSyncCommandCheck(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	os.WriteFile(".env", []byte("FOO=bar\n"), 0644)
	os.WriteFile(".env.example", []byte("BAZ=qux\n"), 0644)

	opts := &syncOptions{
		envPath:     ".env",
		examplePath: ".env.example",
		format:      "text",
		check:       true,
	}
	var stdout, stderr bytes.Buffer
	err := runSync(&stdout, &stderr, opts)
	if err == nil {
		t.Fatal("expected error in check mode")
	}
	if !strings.Contains(stderr.String(), "Drift detected") {
		t.Errorf("expected drift message, got: %s", stderr.String())
	}
}

func TestSyncCommandUnknownFormat(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	os.WriteFile(".env", []byte("FOO=bar\n"), 0644)
	os.WriteFile(".env.example", []byte("BAZ=qux\n"), 0644)

	opts := &syncOptions{
		envPath:     ".env",
		examplePath: ".env.example",
		format:      "xml",
	}
	var stdout, stderr bytes.Buffer
	err := runSync(&stdout, &stderr, opts)
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
}

func TestSyncCommandAddMissing(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	os.WriteFile(".env", []byte("FOO=bar\n"), 0644)
	os.WriteFile(".env.example", []byte("FOO=bar\nBAZ=qux\n"), 0644)

	opts := &syncOptions{
		envPath:     ".env",
		examplePath: ".env.example",
		format:      "text",
		addMissing:  true,
	}
	var stdout, stderr bytes.Buffer
	err := runSync(&stdout, &stderr, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	envContent, _ := os.ReadFile(".env")
	if !strings.Contains(string(envContent), "BAZ=") {
		t.Errorf("expected BAZ to be added to .env")
	}
}

func TestPrintSyncText(t *testing.T) {
	result := &envguardsync.Result{
		Diffs: []envguardsync.Diff{
			{Type: "missing-in-example", Key: "FOO", EnvVal: "bar"},
			{Type: "missing-in-env", Key: "BAZ", ExVal: "qux"},
		},
	}
	var buf bytes.Buffer
	printSyncText(&buf, result, true)
	out := buf.String()
	if !strings.Contains(out, "FOO") {
		t.Errorf("expected FOO in output, got: %s", out)
	}
	if !strings.Contains(out, "BAZ") {
		t.Errorf("expected BAZ in output, got: %s", out)
	}
}

// === watch.go tests ===

func TestWatchCommand(t *testing.T) {
	cmd := newWatchCmd()
	if cmd == nil {
		t.Fatal("expected command")
	}
	if cmd.Use != "watch" {
		t.Errorf("expected use=watch, got: %s", cmd.Use)
	}
}

func TestRunWatchInvalidPaths(t *testing.T) {
	opts := &watchOptions{
		schemaPath: "/nonexistent/schema.yaml",
		envPaths:   []string{"/nonexistent/.env"},
		format:     "text",
		debounce:   50 * time.Millisecond,
		quiet:      true,
	}
	var stdout, stderr bytes.Buffer
	err := runWatch(&stdout, &stderr, opts)
	if err == nil {
		t.Fatal("expected error for invalid paths")
	}
}

// === lsp.go tests ===

func TestLSPCmd(t *testing.T) {
	cmd := newLSPCmd()
	if cmd == nil {
		t.Fatal("expected command")
	}
	if cmd.Use != "lsp" {
		t.Errorf("expected use=lsp, got: %s", cmd.Use)
	}
}

// === root.go tests ===

func TestExecute(t *testing.T) {
	rootCmd.SetArgs([]string{"version"})
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	err := Execute("1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "1.0.0") {
		t.Errorf("expected version output, got: %s", buf.String())
	}
}

// === validate.go additional tests ===

func TestValidateCommandSARIF(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  FOO:
    type: string
    required: true
`), 0644)
	os.WriteFile(envPath, []byte("FOO=bar\n"), 0644)

	cmd := newValidateCmd()
	cmd.SetArgs([]string{"--schema", schemaPath, "--env", envPath, "--format", "sarif"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), `"$schema"`) {
		t.Errorf("expected SARIF output, got: %s", buf.String())
	}
}

func TestValidateCommandFailOnWarnings(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  FOO:
    type: string
    deprecated: "old var"
`), 0644)
	os.WriteFile(envPath, []byte("FOO=bar\n"), 0644)

	cmd := newValidateCmd()
	cmd.SetArgs([]string{"--schema", schemaPath, "--env", envPath, "--fail-on-warnings"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Logf("stdout: %s", buf.String())
		t.Logf("stderr: %s", buf.String())
		t.Fatal("expected error for warnings as errors")
	}
}

func TestValidateCommandStrictMode(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  FOO:
    type: string
`), 0644)
	os.WriteFile(envPath, []byte("FOO=bar\nEXTRA=val\n"), 0644)

	cmd := newValidateCmd()
	cmd.SetArgs([]string{"--schema", schemaPath, "--env", envPath, "--strict"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	// Strict mode currently generates warnings but doesn't fail validation
	_ = err
	if !strings.Contains(buf.String(), "strict") {
		t.Fatalf("expected strict warning in output, got: %s", buf.String())
	}
}

func TestValidateCommandConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  FOO:
    type: string
    required: true
`), 0644)
	os.WriteFile(envPath, []byte("FOO=bar\n"), 0644)
	os.WriteFile(".envguardrc.yaml", []byte(`schema: envguard.yaml
env:
  - .env
format: json
`), 0644)

	cmd := newValidateCmd()
	cmd.SetArgs([]string{})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), `"valid": true`) {
		t.Errorf("expected JSON output from config, got: %s", buf.String())
	}
}

func TestValidateCommandUnknownFormat(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  FOO:
    type: string
`), 0644)
	os.WriteFile(envPath, []byte("FOO=bar\n"), 0644)

	cmd := newValidateCmd()
	cmd.SetArgs([]string{"--schema", schemaPath, "--env", envPath, "--format", "xml"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
}

func TestValidateCommandConfigError(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "bad-config.yaml")
	os.WriteFile(configPath, []byte("not: valid: yaml: {"), 0644)

	opts := &validateOptions{
		configPath: configPath,
	}
	var stdout, stderr bytes.Buffer
	err := runValidate(&stdout, &stderr, opts)
	if err == nil {
		t.Fatal("expected error for bad config")
	}
}

func TestValidateCommandExpandError(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  FOO:
    type: string
`), 0644)
	os.WriteFile(envPath, []byte("FOO=${BAR}\nBAR=${FOO}\n"), 0644)

	cmd := newValidateCmd()
	cmd.SetArgs([]string{"--schema", schemaPath, "--env", envPath})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for circular reference")
	}
}

func TestLoadValidateConfig(t *testing.T) {
	opts := &validateOptions{
		schemaPath:     "/custom/schema.yaml",
		envPaths:       []string{".env", ".env.local"},
		format:         "json",
		strict:         true,
		envName:        "production",
		scanSecrets:    true,
		failOnWarnings: true,
	}
	cfg, err := loadValidateConfig(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Schema != "/custom/schema.yaml" {
		t.Errorf("expected schema override, got: %s", cfg.Schema)
	}
	if cfg.Format != "json" {
		t.Errorf("expected format override, got: %s", cfg.Format)
	}
	if !cfg.Strict {
		t.Error("expected strict=true")
	}
	if cfg.EnvName != "production" {
		t.Errorf("expected envName=production, got: %s", cfg.EnvName)
	}
	if !cfg.ScanSecrets {
		t.Error("expected scanSecrets=true")
	}
	if !cfg.FailOnWarnings {
		t.Error("expected failOnWarnings=true")
	}
}

// === scan.go additional tests ===

func TestScanCommandSARIF(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte("AWS_KEY=AKIAIOSFODNN7EXAMPLE\n"), 0644)

	cmd := newScanCmd()
	cmd.SetArgs([]string{"--env", envPath, "--format", "sarif"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for secrets found")
	}
	if !strings.Contains(buf.String(), `"$schema"`) {
		t.Errorf("expected SARIF output, got: %s", buf.String())
	}
}

func TestScanCommandBaseline(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	baselinePath := filepath.Join(tmpDir, "baseline.json")
	os.WriteFile(envPath, []byte("AWS_KEY=AKIAIOSFODNN7EXAMPLE\n"), 0644)
	os.WriteFile(baselinePath, []byte(`[
        {"key": "AWS_KEY", "rule": "aws-access-key", "severity": "high"}
    ]`), 0644)

	cmd := newScanCmd()
	cmd.SetArgs([]string{"--env", envPath, "--baseline", baselinePath})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "No secrets detected") {
		t.Errorf("expected clean scan after baseline filter, got: %s", buf.String())
	}
}

func TestScanCommandBaselineError(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	baselinePath := filepath.Join(tmpDir, "bad-baseline.json")
	os.WriteFile(envPath, []byte("AWS_KEY=AKIAIOSFODNN7EXAMPLE\n"), 0644)
	os.WriteFile(baselinePath, []byte("not json"), 0644)

	cmd := newScanCmd()
	cmd.SetArgs([]string{"--env", envPath, "--baseline", baselinePath})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for bad baseline")
	}
}

func TestLoadBaseline(t *testing.T) {
	tmpDir := t.TempDir()
	baselinePath := filepath.Join(tmpDir, "baseline.json")
	os.WriteFile(baselinePath, []byte(`[
        {"key": "K", "rule": "R", "severity": "high"}
    ]`), 0644)

	entries, err := loadBaseline(baselinePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got: %d", len(entries))
	}
	if entries[0].Key != "K" {
		t.Errorf("expected key K, got: %s", entries[0].Key)
	}
}

func TestFilterBaseline(t *testing.T) {
	matches := []secrets.SecretMatch{
		{Key: "K1", Rule: "R1"},
		{Key: "K2", Rule: "R2"},
	}
	baseline := []baselineEntry{
		{Key: "K1", Rule: "R1"},
	}
	filtered := filterBaseline(matches, baseline)
	if len(filtered) != 1 {
		t.Fatalf("expected 1 match after filter, got: %d", len(filtered))
	}
	if filtered[0].Key != "K2" {
		t.Errorf("expected K2, got: %s", filtered[0].Key)
	}
}

// === lint.go additional tests ===

func TestLintCommandSARIF(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  FOO:
    type: string
    required: true
    default: "bar"
`), 0644)

	cmd := newLintCmd()
	cmd.SetArgs([]string{"--schema", schemaPath, "--format", "sarif"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(buf.String(), `"$schema"`) {
		t.Errorf("expected SARIF output, got: %s", buf.String())
	}
}

func TestLintCommandUnknownFormat(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  FOO:
    type: string
    required: true
    default: "bar"
`), 0644)

	cmd := newLintCmd()
	cmd.SetArgs([]string{"--schema", schemaPath, "--format", "xml"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
}

func TestLintSchemaAllRules(t *testing.T) {
	allowEmpty := false
	minLen := 10
	maxLen := 5
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"REDUNDANT_REQ": {
				Required: true,
				Default:  "foo",
			},
			"REDUNDANT_EMPTY": {
				Required:   true,
				AllowEmpty: &allowEmpty,
			},
			"UNREACHABLE": {
				DependsOn: "MISSING",
			},
			"EMPTY_ENUM": {
				Enum: []any{},
			},
			"PATTERN_INT": {
				Type:    schema.TypeInteger,
				Pattern: "^[0-9]+$",
			},
			"NO_DESC": {
				Type: schema.TypeString,
			},
			"SUSPICIOUS": {
				Type:    schema.TypeString,
				Default: "changeme",
			},
			"BAD_RANGE": {
				Type: schema.TypeInteger,
				Min:  100,
				Max:  10,
			},
			"BAD_LENGTH": {
				Type:      schema.TypeString,
				MinLength: &minLen,
				MaxLength: &maxLen,
			},
			"DEPRECATED_NO_REPLACEMENT": {
				Type:       schema.TypeString,
				Deprecated: "this is old",
			},
		},
	}
	findings := lintSchema(s)
	if len(findings) == 0 {
		t.Fatal("expected findings")
	}
	rulesFound := make(map[string]bool)
	for _, f := range findings {
		rulesFound[f.Rule] = true
	}
	expectedRules := []string{"redundant", "unreachable", "empty-enum", "type-mismatch", "missing-description", "suspicious-default", "range", "deprecated"}
	for _, rule := range expectedRules {
		if !rulesFound[rule] {
			t.Errorf("expected rule %q to be triggered", rule)
		}
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		input    any
		expected float64
		ok       bool
	}{
		{int(42), 42, true},
		{int8(42), 42, true},
		{int16(42), 42, true},
		{int32(42), 42, true},
		{int64(42), 42, true},
		{uint(42), 42, true},
		{uint8(42), 42, true},
		{uint16(42), 42, true},
		{uint32(42), 42, true},
		{uint64(42), 42, true},
		{float32(42), 42, true},
		{float64(42), 42, true},
		{"string", 0, false},
		{nil, 0, false},
	}
	for _, tt := range tests {
		val, ok := toFloat64(tt.input)
		if ok != tt.ok {
			t.Errorf("toFloat64(%v) ok=%v, want %v", tt.input, ok, tt.ok)
		}
		if ok && val != tt.expected {
			t.Errorf("toFloat64(%v)=%v, want %v", tt.input, val, tt.expected)
		}
	}
}

// === init.go additional tests ===

func TestInitCommandInfer(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	os.WriteFile(".env", []byte("FOO=bar\nPORT=3000\nDEBUG=true\n"), 0644)

	cmd := newInitCmd()
	cmd.SetArgs([]string{"--infer"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "Generated envguard.yaml") {
		t.Errorf("expected generation message, got: %s", buf.String())
	}
	content, _ := os.ReadFile("envguard.yaml")
	if !strings.Contains(string(content), "FOO:") {
		t.Errorf("expected FOO in generated schema")
	}
}

func TestInitCommandConfig(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	cmd := newInitCmd()
	cmd.SetArgs([]string{"--config"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "Created .envguardrc.yaml") {
		t.Errorf("expected creation message, got: %s", buf.String())
	}
}

func TestInitCommandInferMissingEnv(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	cmd := newInitCmd()
	cmd.SetArgs([]string{"--infer", "--env", "nonexistent.env"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing env file")
	}
}

func TestGenerateConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	var stdout, stderr bytes.Buffer
	err := generateConfigFile(&stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "Created .envguardrc.yaml") {
		t.Errorf("expected creation message, got: %s", stdout.String())
	}
	content, _ := os.ReadFile(".envguardrc.yaml")
	if !strings.Contains(string(content), "schema:") {
		t.Errorf("expected schema in config file")
	}
}

// === generate.go additional tests ===

func TestGeneratePlaceholder(t *testing.T) {
	tests := []struct {
		name     string
		v        *schema.Variable
		expected string
	}{
		{
			name:     "default",
			v:        &schema.Variable{Default: "hello"},
			expected: "hello",
		},
		{
			name:     "devOnly",
			v:        &schema.Variable{DevOnly: true},
			expected: "your-dev-value",
		},
		{
			name:     "enum",
			v:        &schema.Variable{Enum: []any{"a", "b"}},
			expected: "a",
		},
		{
			name:     "string_email",
			v:        &schema.Variable{Type: schema.TypeString, Format: "email"},
			expected: "user@example.com",
		},
		{
			name:     "string_url",
			v:        &schema.Variable{Type: schema.TypeString, Format: "url"},
			expected: "https://example.com",
		},
		{
			name:     "string_uuid",
			v:        &schema.Variable{Type: schema.TypeString, Format: "uuid"},
			expected: "00000000-0000-0000-0000-000000000000",
		},
		{
			name:     "string_pattern",
			v:        &schema.Variable{Type: schema.TypeString, Pattern: "^[a-z]+$"},
			expected: "your-value",
		},
		{
			name:     "integer_min",
			v:        &schema.Variable{Type: schema.TypeInteger, Min: 10},
			expected: "10",
		},
		{
			name:     "integer_no_min",
			v:        &schema.Variable{Type: schema.TypeInteger},
			expected: "0",
		},
		{
			name:     "float_min",
			v:        &schema.Variable{Type: schema.TypeFloat, Min: 1.5},
			expected: "1.5",
		},
		{
			name:     "float_no_min",
			v:        &schema.Variable{Type: schema.TypeFloat},
			expected: "0.0",
		},
		{
			name:     "boolean",
			v:        &schema.Variable{Type: schema.TypeBoolean},
			expected: "false",
		},
		{
			name:     "unknown_type",
			v:        &schema.Variable{Type: "unknown"},
			expected: "your-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generatePlaceholder("TEST", tt.v)
			if result != tt.expected {
				t.Errorf("generatePlaceholder()=%q, want %q", result, tt.expected)
			}
		})
	}
}

// === Additional coverage tests ===

func TestDocsCommandUnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  FOO:
    type: string
`), 0644)

	opts := &docsOptions{
		schemaPath: schemaPath,
		format:     "xml",
	}
	var stdout, stderr bytes.Buffer
	err := runDocs(&stdout, &stderr, opts)
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
}

func TestUninstallHookCommandWithArg(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	if err := exec.Command("git", "init").Run(); err != nil {
		t.Skip("git not available")
	}

	opts := &struct {
		hookType string
		force    bool
		command  string
	}{
		hookType: "pre-push",
	}
	var stdout, stderr bytes.Buffer
	runInstallHook(&stdout, &stderr, opts)

	cmd := newUninstallHookCmd()
	cmd.SetArgs([]string{"pre-push"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "Uninstalled pre-push hook") {
		t.Errorf("expected uninstall message, got: %s", buf.String())
	}
}

func TestScanCommandCustomRuleSeverity(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")

	os.WriteFile(envPath, []byte("CUSTOM_TOKEN=iat_abc123def456ghi789jkl012mno345pq\n"), 0644)
	os.WriteFile(schemaPath, []byte(`
version: "1.0"
env:
  CUSTOM_TOKEN:
    type: string
secrets:
  custom:
    - name: internal-api-token
      pattern: "iat_[a-zA-Z0-9]{32}"
      message: "Internal API token detected"
      severity: "critical"
`), 0644)

	cmd := newScanCmd()
	cmd.SetArgs([]string{"--env", envPath, "--schema", schemaPath, "--format", "text"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for custom secret found")
	}
	if !strings.Contains(buf.String(), "[critical]") {
		t.Errorf("expected severity label, got: %s", buf.String())
	}
}

func TestScanCommandSchemaError(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")

	os.WriteFile(envPath, []byte("FOO=bar\n"), 0644)
	os.WriteFile(schemaPath, []byte("invalid yaml: {"), 0644)

	cmd := newScanCmd()
	cmd.SetArgs([]string{"--env", envPath, "--schema", schemaPath})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid schema")
	}
}

func TestSyncCommandSchemaError(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	os.WriteFile(".env", []byte("FOO=bar\n"), 0644)
	os.WriteFile(".env.example", []byte("FOO=bar\n"), 0644)

	opts := &syncOptions{
		envPath:     ".env",
		examplePath: ".env.example",
		schemaPath:  "nonexistent.yaml",
		format:      "text",
	}
	var stdout, stderr bytes.Buffer
	err := runSync(&stdout, &stderr, opts)
	if err == nil {
		t.Fatal("expected error for invalid schema")
	}
}

func TestValidateCommandSchemaError(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte("invalid yaml: {"), 0644)
	os.WriteFile(envPath, []byte("FOO=bar\n"), 0644)

	cmd := newValidateCmd()
	cmd.SetArgs([]string{"--schema", schemaPath, "--env", envPath})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid schema")
	}
}

func TestValidateCommandCustomSecrets(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  API_KEY:
    type: string
    required: true
secrets:
  custom:
    - name: internal-api-token
      pattern: "iat_[a-zA-Z0-9]{32}"
      message: "Internal API token detected"
`), 0644)
	os.WriteFile(envPath, []byte("API_KEY=iat_abc123def456ghi789jkl012mno345pq\n"), 0644)

	cmd := newValidateCmd()
	cmd.SetArgs([]string{"--schema", schemaPath, "--env", envPath, "--scan-secrets"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for secret detected")
	}
	if !strings.Contains(buf.String(), "secret") {
		t.Errorf("expected secret error, got: %s", buf.String())
	}
}

func TestGenerateExampleDevOnlyAndRequiredIn(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	outputPath := filepath.Join(tmpDir, ".env.example")

	schemaContent := `
version: "1.0"
env:
  DEV_VAR:
    type: string
    devOnly: true
    description: "Dev only var"
  PROD_VAR:
    type: string
    requiredIn: ["production"]
    description: "Prod only var"
`
	os.WriteFile(schemaPath, []byte(schemaContent), 0644)

	cmd := newGenerateExampleCmd()
	cmd.SetArgs([]string{"--schema", schemaPath, "--output", outputPath})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("expected output file to exist: %v", err)
	}

	if !strings.Contains(string(content), "DEV_VAR=your-dev-value") {
		t.Errorf("expected devOnly placeholder, got: %s", string(content))
	}
	if !strings.Contains(string(content), "required in: production") {
		t.Errorf("expected requiredIn comment, got: %s", string(content))
	}
}

// === Command execution via cmd.Execute() to cover RunE lines ===

func TestAuditCommandViaExecute(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	envPath := filepath.Join(tmpDir, ".env")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "main.go"), []byte("package main\n"), 0644)
	os.WriteFile(envPath, []byte(""), 0644)

	cmd := newAuditCmd()
	cmd.SetArgs([]string{"--src", srcDir, "--env", envPath})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDocsCommandViaExecute(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  FOO:
    type: string
`), 0644)

	cmd := newDocsCmd()
	cmd.SetArgs([]string{"--schema", schemaPath})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSyncCommandViaExecute(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	os.WriteFile(".env", []byte("FOO=bar\n"), 0644)
	os.WriteFile(".env.example", []byte("FOO=bar\n"), 0644)

	cmd := newSyncCmd()
	cmd.SetArgs([]string{})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWatchCommandViaExecute(t *testing.T) {
	cmd := newWatchCmd()
	cmd.SetArgs([]string{"--schema", "/nonexistent/schema.yaml", "--env", "/nonexistent/.env", "--quiet"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid paths")
	}
}

func TestInstallHookCommandViaExecute(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	if err := exec.Command("git", "init").Run(); err != nil {
		t.Skip("git not available")
	}

	cmd := newInstallHookCmd()
	cmd.SetArgs([]string{})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// === Additional error-path coverage ===

type failingWriter struct{}

func (f failingWriter) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("write failed")
}

func TestLoadBaselineMissingFile(t *testing.T) {
	entries, err := loadBaseline("/nonexistent/baseline.json")
	if err == nil {
		t.Fatal("expected error for missing baseline file")
	}
	if entries != nil {
		t.Error("expected nil entries")
	}
}

func TestValidateCommandJSONEncodeError(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  FOO:
    type: string
    required: true
`), 0644)
	os.WriteFile(envPath, []byte("FOO=bar\n"), 0644)

	opts := &validateOptions{
		schemaPath: schemaPath,
		envPaths:   []string{envPath},
		format:     "json",
	}
	var stdout failingWriter
	var stderr bytes.Buffer
	err := runValidate(&stdout, &stderr, opts)
	if err == nil {
		t.Fatal("expected error for JSON encode failure")
	}
}

func TestValidateCommandSARIFEncodeError(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  FOO:
    type: string
    required: true
`), 0644)
	os.WriteFile(envPath, []byte("FOO=bar\n"), 0644)

	opts := &validateOptions{
		schemaPath: schemaPath,
		envPaths:   []string{envPath},
		format:     "sarif",
	}
	var stdout failingWriter
	var stderr bytes.Buffer
	err := runValidate(&stdout, &stderr, opts)
	if err == nil {
		t.Fatal("expected error for SARIF encode failure")
	}
}

func TestScanCommandJSONEncodeError(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte("AWS_KEY=AKIAIOSFODNN7EXAMPLE\n"), 0644)

	opts := &scanOptions{
		envPaths: []string{envPath},
		format:   "json",
	}
	var stdout failingWriter
	var stderr bytes.Buffer
	err := runScan(&stdout, &stderr, opts)
	if err == nil {
		t.Fatal("expected error for JSON encode failure")
	}
}

func TestScanCommandSARIFEncodeError(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte("AWS_KEY=AKIAIOSFODNN7EXAMPLE\n"), 0644)

	opts := &scanOptions{
		envPaths: []string{envPath},
		format:   "sarif",
	}
	var stdout failingWriter
	var stderr bytes.Buffer
	err := runScan(&stdout, &stderr, opts)
	if err == nil {
		t.Fatal("expected error for SARIF encode failure")
	}
}

func TestSyncCommandJSONEncodeError(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	os.WriteFile(".env", []byte("FOO=bar\n"), 0644)
	os.WriteFile(".env.example", []byte("BAZ=qux\n"), 0644)

	opts := &syncOptions{
		envPath:     ".env",
		examplePath: ".env.example",
		format:      "json",
	}
	var stdout failingWriter
	var stderr bytes.Buffer
	err := runSync(&stdout, &stderr, opts)
	if err == nil {
		t.Fatal("expected error for JSON encode failure")
	}
}

func TestSyncCommandSARIFEncodeError(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	os.WriteFile(".env", []byte("FOO=bar\n"), 0644)
	os.WriteFile(".env.example", []byte("BAZ=qux\n"), 0644)

	opts := &syncOptions{
		envPath:     ".env",
		examplePath: ".env.example",
		format:      "sarif",
	}
	var stdout failingWriter
	var stderr bytes.Buffer
	err := runSync(&stdout, &stderr, opts)
	if err == nil {
		t.Fatal("expected error for SARIF encode failure")
	}
}

func TestAuditCommandJSONEncodeError(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	envPath := filepath.Join(tmpDir, ".env")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "main.go"), []byte(`package main
import "os"
func main() {
    _ = os.Getenv("MISSING_VAR")
}
`), 0644)
	os.WriteFile(envPath, []byte(""), 0644)

	opts := &auditOptions{
		srcDir:  srcDir,
		envPath: envPath,
		format:  "json",
	}
	var stdout failingWriter
	var stderr bytes.Buffer
	err := runAudit(&stdout, &stderr, opts)
	if err == nil {
		t.Fatal("expected error for JSON encode failure")
	}
}

func TestAuditCommandSARIFEncodeError(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	envPath := filepath.Join(tmpDir, ".env")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "main.go"), []byte(`package main
import "os"
func main() {
    _ = os.Getenv("MISSING_VAR")
}
`), 0644)
	os.WriteFile(envPath, []byte(""), 0644)

	opts := &auditOptions{
		srcDir:  srcDir,
		envPath: envPath,
		format:  "sarif",
	}
	var stdout failingWriter
	var stderr bytes.Buffer
	err := runAudit(&stdout, &stderr, opts)
	if err == nil {
		t.Fatal("expected error for SARIF encode failure")
	}
}

// === Final coverage tests ===

func TestLintCommandSchemaParseError(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte("invalid yaml: {"), 0644)

	cmd := newLintCmd()
	cmd.SetArgs([]string{"--schema", schemaPath})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid schema")
	}
}

func TestLintCommandSARIFEncodeError(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  FOO:
    type: string
    required: true
    default: "bar"
`), 0644)

	opts := &lintOptions{
		schemaPath: schemaPath,
		format:     "sarif",
	}
	var stdout failingWriter
	var stderr bytes.Buffer
	err := runLint(&stdout, &stderr, opts)
	if err == nil {
		t.Fatal("expected error for SARIF encode failure")
	}
}

func TestValidateCommandMediumSeveritySecret(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  API_KEY:
    type: string
    required: true
`), 0644)
	os.WriteFile(envPath, []byte("API_KEY=api_key=abcdefghijklmnopqrstuvwxyz1234\n"), 0644)

	cmd := newValidateCmd()
	cmd.SetArgs([]string{"--schema", schemaPath, "--env", envPath, "--scan-secrets"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Medium severity secrets are added as warnings, not errors
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "secret") {
		t.Errorf("expected secret warning in output, got: %s", buf.String())
	}
}

func TestValidateCommandLowSeveritySecret(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")
	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(schemaPath, []byte(`version: "1.0"
env:
  DATA:
    type: string
    required: true
`), 0644)
	os.WriteFile(envPath, []byte("DATA=ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890\n"), 0644)

	cmd := newValidateCmd()
	cmd.SetArgs([]string{"--schema", schemaPath, "--env", envPath, "--scan-secrets"})
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Low severity secrets are added as info, not errors
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "secret") {
		t.Errorf("expected secret info in output, got: %s", buf.String())
	}
}
