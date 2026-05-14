// Package monorepo provides support for validating multiple .env files across a project.
package monorepo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EnvProject represents a discovered .env file with its associated schema.
type EnvProject struct {
	EnvPath    string `json:"envPath"`
	SchemaPath string `json:"schemaPath"`
	Dir        string `json:"dir"`
}

// Discover finds all .env files and their associated schemas recursively.
func Discover(root string, recursive bool) ([]EnvProject, error) {
	var projects []EnvProject

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			name := info.Name()
			if name == "node_modules" || name == "vendor" || name == ".git" || name == "dist" || name == "build" || name == "target" || name == ".venv" || name == "__pycache__" {
				return filepath.SkipDir
			}
			if !recursive && path != root {
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Base(path) == ".env" {
			dir := filepath.Dir(path)
			schemaPath := findSchemaForDir(dir)
			projects = append(projects, EnvProject{
				EnvPath:    path,
				SchemaPath: schemaPath,
				Dir:        dir,
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return projects, nil
}

// findSchemaForDir looks for envguard.yaml in the directory or its parents.
func findSchemaForDir(dir string) string {
	for {
		schemaPath := filepath.Join(dir, "envguard.yaml")
		if _, err := os.Stat(schemaPath); err == nil {
			return schemaPath
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// ValidateAllResult holds the combined validation results.
type ValidateAllResult struct {
	Projects []ProjectResult `json:"projects"`
	Valid    bool            `json:"valid"`
}

// ProjectResult holds validation results for a single project.
type ProjectResult struct {
	Project EnvProject `json:"project"`
	Valid   bool       `json:"valid"`
	Errors  []string   `json:"errors,omitempty"`
}

// FormatResults produces human-readable output from validation results.
func FormatResults(results []ProjectResult) string {
	var b strings.Builder
	validCount := 0
	invalidCount := 0

	for _, r := range results {
		if r.Valid {
			validCount++
			fmt.Fprintf(&b, "  ✓ %s\n", r.Project.Dir)
		} else {
			invalidCount++
			fmt.Fprintf(&b, "  ✗ %s\n", r.Project.Dir)
			for _, err := range r.Errors {
				fmt.Fprintf(&b, "    └─ %s\n", err)
			}
		}
	}

	fmt.Fprintf(&b, "\n%d passed, %d failed\n", validCount, invalidCount)
	return b.String()
}
