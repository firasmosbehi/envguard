// Package config provides loading and merging of EnvGuard configuration files.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds project-wide defaults for EnvGuard.
type Config struct {
	Schema         string   `yaml:"schema,omitempty"`
	Env            []string `yaml:"env,omitempty"`
	Format         string   `yaml:"format,omitempty"`
	Strict         bool     `yaml:"strict,omitempty"`
	EnvName        string   `yaml:"envName,omitempty"`
	ScanSecrets    bool     `yaml:"scanSecrets,omitempty"`
	FailOnWarnings bool     `yaml:"failOnWarnings,omitempty"`
}

// Default returns a Config populated with built-in defaults.
func Default() *Config {
	return &Config{
		Schema: "envguard.yaml",
		Env:    []string{".env"},
		Format: "text",
	}
}

// Find searches for a config file starting from the given directory and
// walking up to the git root, then the user's home directory.
// Supported file names (in order of preference):
//
//	.envguardrc.yaml, .envguardrc.yml, .envguardrc, envguard.config.yaml, envguard.config.yml
func Find(startDir string) (string, bool) {
	names := []string{
		".envguardrc.yaml",
		".envguardrc.yml",
		".envguardrc",
		"envguard.config.yaml",
		"envguard.config.yml",
	}

	// Walk up from startDir
	for dir := startDir; dir != "/" && dir != "."; dir = filepath.Dir(dir) {
		for _, name := range names {
			path := filepath.Join(dir, name)
			if _, err := os.Stat(path); err == nil {
				return path, true
			}
		}
		// Stop at git root
		gitDir := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			break
		}
	}

	// Check home directory
	home, err := os.UserHomeDir()
	if err == nil {
		for _, name := range names {
			path := filepath.Join(home, name)
			if _, err := os.Stat(path); err == nil {
				return path, true
			}
		}
	}

	return "", false
}

// Load reads and parses a config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	return &cfg, nil
}

// Merge combines a base config with an override config.
// Override values take precedence over base values.
func Merge(base, override *Config) *Config {
	merged := &Config{
		Schema:         base.Schema,
		Env:            base.Env,
		Format:         base.Format,
		Strict:         base.Strict,
		EnvName:        base.EnvName,
		ScanSecrets:    base.ScanSecrets,
		FailOnWarnings: base.FailOnWarnings,
	}

	if override.Schema != "" {
		merged.Schema = override.Schema
	}
	if len(override.Env) > 0 {
		merged.Env = override.Env
	}
	if override.Format != "" {
		merged.Format = override.Format
	}
	merged.Strict = override.Strict || base.Strict
	if override.EnvName != "" {
		merged.EnvName = override.EnvName
	}
	merged.ScanSecrets = override.ScanSecrets || base.ScanSecrets
	merged.FailOnWarnings = override.FailOnWarnings || base.FailOnWarnings

	return merged
}

// EnvOverride applies environment-variable overrides to a config.
// Supported env vars: ENGUARD_SCHEMA, ENGUARD_ENV, ENGUARD_FORMAT,
// ENGUARD_STRICT, ENGUARD_ENV_NAME, ENGUARD_SCAN_SECRETS.
func EnvOverride(cfg *Config) *Config {
	if v := os.Getenv("ENGUARD_SCHEMA"); v != "" {
		cfg.Schema = v
	}
	if v := os.Getenv("ENGUARD_ENV"); v != "" {
		cfg.Env = strings.Split(v, ",")
	}
	if v := os.Getenv("ENGUARD_FORMAT"); v != "" {
		cfg.Format = v
	}
	if v := os.Getenv("ENGUARD_STRICT"); v == "true" || v == "1" {
		cfg.Strict = true
	}
	if v := os.Getenv("ENGUARD_ENV_NAME"); v != "" {
		cfg.EnvName = v
	}
	if v := os.Getenv("ENGUARD_SCAN_SECRETS"); v == "true" || v == "1" {
		cfg.ScanSecrets = true
	}
	if v := os.Getenv("ENGUARD_FAIL_ON_WARNINGS"); v == "true" || v == "1" {
		cfg.FailOnWarnings = true
	}
	return cfg
}
