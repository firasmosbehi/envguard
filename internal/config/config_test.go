package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Schema != "envguard.yaml" {
		t.Errorf("expected schema 'envguard.yaml', got %q", cfg.Schema)
	}
	if len(cfg.Env) != 1 || cfg.Env[0] != ".env" {
		t.Errorf("expected env ['.env'], got %v", cfg.Env)
	}
	if cfg.Format != "text" {
		t.Errorf("expected format 'text', got %q", cfg.Format)
	}
}

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".envguardrc.yaml")
	content := `
schema: config-schema.yaml
env:
  - .env
  - .env.local
format: json
strict: true
envName: production
scanSecrets: true
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Schema != "config-schema.yaml" {
		t.Errorf("expected schema 'config-schema.yaml', got %q", cfg.Schema)
	}
	if len(cfg.Env) != 2 {
		t.Errorf("expected 2 env files, got %d", len(cfg.Env))
	}
	if cfg.Format != "json" {
		t.Errorf("expected format 'json', got %q", cfg.Format)
	}
	if !cfg.Strict {
		t.Error("expected strict=true")
	}
	if cfg.EnvName != "production" {
		t.Errorf("expected envName 'production', got %q", cfg.EnvName)
	}
	if !cfg.ScanSecrets {
		t.Error("expected scanSecrets=true")
	}
}

func TestLoadNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/.envguardrc.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, ".envguardrc.yaml")
	content := `
{
  invalid yaml: [
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestMerge(t *testing.T) {
	base := &Config{
		Schema:  "base.yaml",
		Env:     []string{".env"},
		Format:  "text",
		Strict:  false,
		EnvName: "dev",
	}
	override := &Config{
		Schema:  "override.yaml",
		Format:  "json",
		Strict:  true,
		EnvName: "",
	}

	merged := Merge(base, override)

	if merged.Schema != "override.yaml" {
		t.Errorf("expected schema 'override.yaml', got %q", merged.Schema)
	}
	if len(merged.Env) != 1 || merged.Env[0] != ".env" {
		t.Errorf("expected env from base, got %v", merged.Env)
	}
	if merged.Format != "json" {
		t.Errorf("expected format 'json', got %q", merged.Format)
	}
	if !merged.Strict {
		t.Error("expected strict=true from override")
	}
	if merged.EnvName != "dev" {
		t.Errorf("expected envName 'dev' from base, got %q", merged.EnvName)
	}
}

func TestMergeAllFields(t *testing.T) {
	base := &Config{
		Schema:         "base.yaml",
		Env:            []string{".env"},
		Format:         "text",
		Strict:         false,
		EnvName:        "dev",
		ScanSecrets:    false,
		FailOnWarnings: false,
	}
	override := &Config{
		Schema:         "override.yaml",
		Env:            []string{".env", ".env.local"},
		Format:         "json",
		Strict:         true,
		EnvName:        "prod",
		ScanSecrets:    true,
		FailOnWarnings: true,
	}

	merged := Merge(base, override)

	if merged.Schema != "override.yaml" {
		t.Errorf("expected schema 'override.yaml', got %q", merged.Schema)
	}
	if len(merged.Env) != 2 {
		t.Errorf("expected 2 env files, got %d", len(merged.Env))
	}
	if merged.Format != "json" {
		t.Errorf("expected format 'json', got %q", merged.Format)
	}
	if !merged.Strict {
		t.Error("expected strict=true")
	}
	if merged.EnvName != "prod" {
		t.Errorf("expected envName 'prod', got %q", merged.EnvName)
	}
	if !merged.ScanSecrets {
		t.Error("expected scanSecrets=true")
	}
	if !merged.FailOnWarnings {
		t.Error("expected failOnWarnings=true")
	}
}

func TestMergeEmptyOverride(t *testing.T) {
	base := &Config{
		Schema:         "base.yaml",
		Env:            []string{".env"},
		Format:         "text",
		Strict:         true,
		EnvName:        "dev",
		ScanSecrets:    true,
		FailOnWarnings: true,
	}
	override := &Config{}

	merged := Merge(base, override)

	if merged.Schema != "base.yaml" {
		t.Errorf("expected schema 'base.yaml', got %q", merged.Schema)
	}
	if !merged.Strict {
		t.Error("expected strict=true from base")
	}
	if merged.EnvName != "dev" {
		t.Errorf("expected envName 'dev', got %q", merged.EnvName)
	}
	if !merged.ScanSecrets {
		t.Error("expected scanSecrets=true from base")
	}
	if !merged.FailOnWarnings {
		t.Error("expected failOnWarnings=true from base")
	}
}

func TestMergeBaseTrueOverrideFalse(t *testing.T) {
	// Test that boolean OR works correctly when base is true and override is false
	base := &Config{
		Strict:         true,
		ScanSecrets:    true,
		FailOnWarnings: true,
	}
	override := &Config{
		Strict:         false,
		ScanSecrets:    false,
		FailOnWarnings: false,
	}

	merged := Merge(base, override)
	if !merged.Strict {
		t.Error("expected strict=true (base true OR override false)")
	}
	if !merged.ScanSecrets {
		t.Error("expected scanSecrets=true")
	}
	if !merged.FailOnWarnings {
		t.Error("expected failOnWarnings=true")
	}
}

func TestFind(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create config in tmpDir
	configPath := filepath.Join(tmpDir, ".envguardrc.yaml")
	if err := os.WriteFile(configPath, []byte("format: json\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Should find from subDir
	found, ok := Find(subDir)
	if !ok {
		t.Error("expected to find config file")
	}
	if found != configPath {
		t.Errorf("expected path %q, got %q", configPath, found)
	}
}

func TestFindWithGitRoot(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create .git directory to act as root
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create config inside git root
	configPath := filepath.Join(tmpDir, ".envguardrc.yaml")
	if err := os.WriteFile(configPath, []byte("format: json\n"), 0644); err != nil {
		t.Fatal(err)
	}

	found, ok := Find(subDir)
	if !ok {
		t.Error("expected to find config file")
	}
	if found != configPath {
		t.Errorf("expected path %q, got %q", configPath, found)
	}
}

func TestFindInHomeDir(t *testing.T) {
	// Create a temp home directory
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	configPath := filepath.Join(tmpHome, ".envguardrc.yaml")
	if err := os.WriteFile(configPath, []byte("format: json\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Find from a directory with no config and no git root
	searchDir := filepath.Join(tmpHome, "no-config-here")
	os.MkdirAll(searchDir, 0755)

	found, ok := Find(searchDir)
	if !ok {
		t.Error("expected to find config file in home dir")
	}
	if found != configPath {
		t.Errorf("expected path %q, got %q", configPath, found)
	}
}

func TestFindNotFound(t *testing.T) {
	// Create a temp home directory with no config
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	searchDir := filepath.Join(tmpHome, "no-config")
	os.MkdirAll(searchDir, 0755)

	found, ok := Find(searchDir)
	if ok {
		t.Error("expected not to find config file")
	}
	if found != "" {
		t.Errorf("expected empty path, got %q", found)
	}
}

func TestFindMultipleNames(t *testing.T) {
	tmpDir := t.TempDir()
	// Test that .envguardrc.yml is found
	configPath := filepath.Join(tmpDir, ".envguardrc.yml")
	os.WriteFile(configPath, []byte("format: json\n"), 0644)

	found, ok := Find(tmpDir)
	if !ok {
		t.Fatal("expected to find config file")
	}
	if found != configPath {
		t.Errorf("expected path %q, got %q", configPath, found)
	}
}

func TestFindEnvguardConfigYaml(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "envguard.config.yaml")
	os.WriteFile(configPath, []byte("format: json\n"), 0644)

	found, ok := Find(tmpDir)
	if !ok {
		t.Fatal("expected to find config file")
	}
	if found != configPath {
		t.Errorf("expected path %q, got %q", configPath, found)
	}
}

func TestEnvOverride(t *testing.T) {
	t.Setenv("ENGUARD_SCHEMA", "env-schema.yaml")
	t.Setenv("ENGUARD_FORMAT", "json")
	t.Setenv("ENGUARD_STRICT", "true")
	t.Setenv("ENGUARD_ENV_NAME", "staging")
	t.Setenv("ENGUARD_SCAN_SECRETS", "1")

	cfg := Default()
	cfg = EnvOverride(cfg)

	if cfg.Schema != "env-schema.yaml" {
		t.Errorf("expected schema 'env-schema.yaml', got %q", cfg.Schema)
	}
	if cfg.Format != "json" {
		t.Errorf("expected format 'json', got %q", cfg.Format)
	}
	if !cfg.Strict {
		t.Error("expected strict=true")
	}
	if cfg.EnvName != "staging" {
		t.Errorf("expected envName 'staging', got %q", cfg.EnvName)
	}
	if !cfg.ScanSecrets {
		t.Error("expected scanSecrets=true")
	}
}

func TestEnvOverrideAllVars(t *testing.T) {
	t.Setenv("ENGUARD_SCHEMA", "s.yaml")
	t.Setenv("ENGUARD_ENV", ".env,.env.local")
	t.Setenv("ENGUARD_FORMAT", "json")
	t.Setenv("ENGUARD_STRICT", "true")
	t.Setenv("ENGUARD_ENV_NAME", "prod")
	t.Setenv("ENGUARD_SCAN_SECRETS", "1")
	t.Setenv("ENGUARD_FAIL_ON_WARNINGS", "true")

	cfg := Default()
	cfg = EnvOverride(cfg)

	if cfg.Schema != "s.yaml" {
		t.Errorf("expected schema 's.yaml', got %q", cfg.Schema)
	}
	if len(cfg.Env) != 2 || cfg.Env[0] != ".env" || cfg.Env[1] != ".env.local" {
		t.Errorf("expected env ['.env', '.env.local'], got %v", cfg.Env)
	}
	if cfg.Format != "json" {
		t.Errorf("expected format 'json', got %q", cfg.Format)
	}
	if !cfg.Strict {
		t.Error("expected strict=true")
	}
	if cfg.EnvName != "prod" {
		t.Errorf("expected envName 'prod', got %q", cfg.EnvName)
	}
	if !cfg.ScanSecrets {
		t.Error("expected scanSecrets=true")
	}
	if !cfg.FailOnWarnings {
		t.Error("expected failOnWarnings=true")
	}
}

func TestEnvOverrideNoVars(t *testing.T) {
	// Ensure no ENGUARD_* env vars are set
	for _, key := range []string{
		"ENGUARD_SCHEMA", "ENGUARD_ENV", "ENGUARD_FORMAT",
		"ENGUARD_STRICT", "ENGUARD_ENV_NAME", "ENGUARD_SCAN_SECRETS",
		"ENGUARD_FAIL_ON_WARNINGS",
	} {
		os.Unsetenv(key)
	}

	cfg := Default()
	cfg = EnvOverride(cfg)

	if cfg.Schema != "envguard.yaml" {
		t.Errorf("expected default schema, got %q", cfg.Schema)
	}
	if cfg.Format != "text" {
		t.Errorf("expected default format, got %q", cfg.Format)
	}
	if cfg.Strict {
		t.Error("expected strict=false")
	}
	if cfg.ScanSecrets {
		t.Error("expected scanSecrets=false")
	}
	if cfg.FailOnWarnings {
		t.Error("expected failOnWarnings=false")
	}
}

func TestEnvOverrideStrictFalse(t *testing.T) {
	// Test that values other than "true" or "1" don't set strict to true,
	// but also don't reset an already-true value to false.
	t.Setenv("ENGUARD_STRICT", "false")
	t.Setenv("ENGUARD_SCAN_SECRETS", "0")
	t.Setenv("ENGUARD_FAIL_ON_WARNINGS", "no")

	cfg := Default()
	// Default has false values; override with non-true strings should keep them false
	cfg = EnvOverride(cfg)

	if cfg.Strict {
		t.Error("expected strict=false")
	}
	if cfg.ScanSecrets {
		t.Error("expected scanSecrets=false")
	}
	if cfg.FailOnWarnings {
		t.Error("expected failOnWarnings=false")
	}
}
