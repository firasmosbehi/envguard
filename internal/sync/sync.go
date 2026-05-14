// Package sync provides bidirectional sync between .env and .env.example files.
package sync

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/envguard/envguard/internal/dotenv"
	"github.com/envguard/envguard/internal/schema"
)

// Diff represents a single difference between .env and .env.example.
type Diff struct {
	Type   string `json:"type"` // "missing-in-example", "missing-in-env", "value-mismatch"
	Key    string `json:"key"`
	EnvVal string `json:"envValue,omitempty"`
	ExVal  string `json:"exampleValue,omitempty"`
}

// Result holds the sync analysis output.
type Result struct {
	Diffs      []Diff `json:"diffs"`
	WouldWrite bool   `json:"wouldWrite"`
}

// Options configures the sync behavior.
type Options struct {
	EnvPath     string
	ExamplePath string
	SchemaPath  string
	Check       bool // dry-run / CI mode
	AddMissing  bool // add missing keys to .env
}

// envEntry represents a parsed line from an .env file.
type envEntry struct {
	Key     string
	Value   string
	Comment string
	Raw     string
	IsBlank bool
}

// parseEnvFile parses an .env file preserving comments and order.
func parseEnvFile(path string) ([]envEntry, map[string]envEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, make(map[string]envEntry), nil
		}
		return nil, nil, err
	}
	defer file.Close()

	var entries []envEntry
	lookup := make(map[string]envEntry)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			entries = append(entries, envEntry{Raw: line, IsBlank: true})
			continue
		}

		if strings.HasPrefix(trimmed, "#") {
			entries = append(entries, envEntry{Raw: line, Comment: trimmed})
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove quotes if present
			if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
				value = value[1 : len(value)-1]
			}
			entry := envEntry{Key: key, Value: value, Raw: line}
			entries = append(entries, entry)
			lookup[key] = entry
		} else {
			entries = append(entries, envEntry{Raw: line})
		}
	}

	return entries, lookup, scanner.Err()
}

// Run performs the sync analysis and optionally writes the example file.
func Run(opts Options) (*Result, error) {
	envEntries, envLookup, err := parseEnvFile(opts.EnvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse .env: %w", err)
	}

	exampleEntries, exampleLookup, err := parseEnvFile(opts.ExamplePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse .env.example: %w", err)
	}

	var sch *schema.Schema
	if opts.SchemaPath != "" {
		s, err := schema.Parse(opts.SchemaPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse schema: %w", err)
		}
		sch = s
	}

	result := &Result{Diffs: make([]Diff, 0)}

	// Build sensitive set from schema
	sensitiveSet := make(map[string]bool)
	if sch != nil {
		for name, v := range sch.Env {
			if v.Sensitive {
				sensitiveSet[name] = true
			}
		}
	}

	// Check: keys in .env but not in .env.example
	for key := range envLookup {
		if _, exists := exampleLookup[key]; !exists {
			val := envLookup[key].Value
			if sensitiveSet[key] {
				val = maskValue(val)
			}
			result.Diffs = append(result.Diffs, Diff{
				Type:   "missing-in-example",
				Key:    key,
				EnvVal: val,
			})
		}
	}

	// Check: keys in .env.example but not in .env
	for key := range exampleLookup {
		if _, exists := envLookup[key]; !exists {
			result.Diffs = append(result.Diffs, Diff{
				Type:  "missing-in-env",
				Key:   key,
				ExVal: exampleLookup[key].Value,
			})
		}
	}

	// Check: schema-defined variables not in either file
	if sch != nil {
		for name := range sch.Env {
			if _, inEnv := envLookup[name]; !inEnv {
				if _, inEx := exampleLookup[name]; !inEx {
					result.Diffs = append(result.Diffs, Diff{
						Type: "missing-in-example",
						Key:  name,
					})
				}
			}
		}
	}

	if opts.Check {
		result.WouldWrite = len(result.Diffs) > 0
		return result, nil
	}

	// Generate the new .env.example content
	newExample := generateExample(envEntries, exampleEntries, envLookup, exampleLookup, sensitiveSet, sch)

	// Check if there's actual change
	existingContent, _ := os.ReadFile(opts.ExamplePath)
	if string(existingContent) == newExample {
		result.WouldWrite = false
		return result, nil
	}

	result.WouldWrite = true

	if !opts.Check {
		if err := os.WriteFile(opts.ExamplePath, []byte(newExample), 0644); err != nil {
			return nil, fmt.Errorf("failed to write .env.example: %w", err)
		}
	}

	// If --add-missing, add missing keys to .env
	if opts.AddMissing {
		for _, diff := range result.Diffs {
			if diff.Type == "missing-in-env" {
				f, err := os.OpenFile(opts.EnvPath, os.O_APPEND|os.O_WRONLY, 0644)
				if err != nil {
					return nil, fmt.Errorf("failed to open .env: %w", err)
				}
				fmt.Fprintf(f, "\n%s=", diff.Key)
				f.Close()
			}
		}
	}

	return result, nil
}

// generateExample creates the content for .env.example based on .env and schema.
func generateExample(envEntries []envEntry, exampleEntries []envEntry, envLookup, exampleLookup map[string]envEntry, sensitiveSet map[string]bool, sch *schema.Schema) string {
	var lines []string
	seen := make(map[string]bool)

	// Add a header comment
	lines = append(lines, "# Auto-generated by EnvGuard sync")
	lines = append(lines, "# Copy this file to .env and fill in the values")
	lines = append(lines, "")

	// First, include all keys from .env in their original order
	for _, entry := range envEntries {
		if entry.Key == "" {
			if entry.IsBlank && len(lines) > 0 && lines[len(lines)-1] == "" {
				continue // skip consecutive blanks
			}
			if entry.Comment != "" {
				lines = append(lines, entry.Raw)
			} else if entry.IsBlank {
				lines = append(lines, "")
			}
			continue
		}
		if seen[entry.Key] {
			continue
		}
		seen[entry.Key] = true

		val := entry.Value
		if sensitiveSet[entry.Key] {
			val = maskValue(val)
		} else if val != "" {
			val = suggestPlaceholder(entry.Key, val, sch)
		}

		if val == "" {
			lines = append(lines, fmt.Sprintf("%s=", entry.Key))
		} else {
			lines = append(lines, fmt.Sprintf("%s=%s", entry.Key, val))
		}
	}

	// Then add any schema-defined variables not yet included
	if sch != nil {
		var schemaKeys []string
		for name := range sch.Env {
			if !seen[name] {
				schemaKeys = append(schemaKeys, name)
			}
		}
		sort.Strings(schemaKeys)
		if len(schemaKeys) > 0 {
			lines = append(lines, "")
			lines = append(lines, "# Schema-defined variables")
			for _, key := range schemaKeys {
				seen[key] = true
				var val string
				if sensitiveSet[key] {
					val = "***"
				} else {
					val = suggestPlaceholder(key, "", sch)
				}
				if val == "" {
					lines = append(lines, fmt.Sprintf("%s=", key))
				} else {
					lines = append(lines, fmt.Sprintf("%s=%s", key, val))
				}
			}
		}
	}

	// Trim trailing blank lines
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return strings.Join(lines, "\n") + "\n"
}

// maskValue replaces a sensitive value with a placeholder.
func maskValue(value string) string {
	if value == "" {
		return ""
	}
	return "***"
}

// suggestPlaceholder generates a smart placeholder based on the variable name and schema.
func suggestPlaceholder(key, currentValue string, sch *schema.Schema) string {
	if currentValue != "" {
		// Try to preserve structure for common patterns
		lower := strings.ToLower(key)
		if strings.Contains(lower, "url") || strings.Contains(lower, "uri") || strings.Contains(lower, "host") {
			if strings.HasPrefix(currentValue, "http://") || strings.HasPrefix(currentValue, "https://") {
				return currentValue
			}
			if strings.Contains(lower, "database") || strings.Contains(lower, "db") || strings.Contains(lower, "postgres") || strings.Contains(lower, "mysql") {
				return "postgresql://user:password@localhost:5432/dbname"
			}
			if strings.Contains(lower, "redis") {
				return "redis://localhost:6379"
			}
			return "https://example.com"
		}
		if strings.Contains(lower, "port") {
			return currentValue
		}
		if strings.Contains(lower, "key") || strings.Contains(lower, "token") || strings.Contains(lower, "secret") {
			return "your-" + strings.ReplaceAll(lower, "_", "-") + "-here"
		}
		return currentValue
	}

	// For empty values, use schema defaults if available
	if sch != nil {
		if v, ok := sch.Env[key]; ok && v.Default != nil {
			return fmt.Sprintf("%v", v.Default)
		}
	}

	return ""
}

// ParseEnv wraps dotenv.Parse for use by sync.
func ParseEnv(path string) (map[string]string, error) {
	return dotenv.Parse(path)
}
