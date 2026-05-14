// Package infer analyzes .env files to infer schema types.
package infer

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/envguard/envguard/internal/schema"
)

// Result holds inferred schema variables.
type Result struct {
	Variables map[string]*schema.Variable
}

// FromEnv analyzes env vars and infers their schema definitions.
func FromEnv(envVars map[string]string) *Result {
	result := &Result{Variables: make(map[string]*schema.Variable)}
	for key, value := range envVars {
		result.Variables[key] = inferVariable(key, value)
	}
	return result
}

func inferVariable(key, value string) *schema.Variable {
	v := &schema.Variable{
		Type:        schema.TypeString,
		Description: inferDescription(key),
	}

	// Try boolean
	if isBoolean(value) {
		v.Type = schema.TypeBoolean
		if bv, _ := strconv.ParseBool(value); bv {
			v.Default = true
		} else {
			v.Default = false
		}
		return v
	}

	// Try integer
	if i, err := strconv.ParseInt(value, 10, 64); err == nil {
		v.Type = schema.TypeInteger
		v.Default = int(i)
		// Common integer patterns
		lower := strings.ToLower(key)
		if strings.Contains(lower, "port") {
			v.Min = 1
			v.Max = 65535
		}
		return v
	}

	// Try float
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		v.Type = schema.TypeFloat
		v.Default = f
		return v
	}

	// Try array
	if strings.Contains(value, ",") && !strings.Contains(value, "//") {
		parts := strings.Split(value, ",")
		if len(parts) > 1 {
			v.Type = schema.TypeArray
			v.Separator = ","
			return v
		}
	}

	// String type with format detection
	v.Format = inferFormat(key, value)

	// Pattern detection for common string types
	lower := strings.ToLower(key)
	if strings.Contains(lower, "url") || strings.Contains(lower, "uri") || strings.Contains(lower, "endpoint") {
		v.Format = "url"
	} else if strings.Contains(lower, "host") && !strings.Contains(lower, "@") {
		v.Format = "hostname"
	} else if strings.Contains(lower, "email") {
		v.Format = "email"
	} else if strings.Contains(lower, "key") || strings.Contains(lower, "secret") || strings.Contains(lower, "token") || strings.Contains(lower, "password") {
		v.Sensitive = true
	}

	return v
}

func isBoolean(s string) bool {
	switch strings.ToLower(s) {
	case "true", "false", "1", "0", "yes", "no", "on", "off":
		return true
	}
	return false
}

func inferFormat(key, value string) string {
	// UUID
	if matched, _ := regexp.MatchString(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, strings.ToLower(value)); matched {
		return "uuid"
	}
	// Email
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, value); matched {
		return "email"
	}
	// URL
	if u, err := url.Parse(value); err == nil && (u.Scheme == "http" || u.Scheme == "https" || u.Scheme == "redis" || u.Scheme == "postgresql" || u.Scheme == "mongodb" || u.Scheme == "mysql") {
		return "url"
	}
	// IP
	if matched, _ := regexp.MatchString(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`, value); matched {
		return "ip"
	}
	// Hex
	if matched, _ := regexp.MatchString(`^[0-9a-fA-F]+$`, value); matched && len(value) >= 8 {
		return "hex"
	}
	// Semver
	if matched, _ := regexp.MatchString(`^v?\d+\.\d+\.\d+`, value); matched {
		return "semver"
	}
	// Duration
	if matched, _ := regexp.MatchString(`^\d+(ns|us|µs|ms|s|m|h)$`, value); matched {
		return "duration"
	}
	// Base64
	if matched, _ := regexp.MatchString(`^[A-Za-z0-9+/=]+$`, value); matched && len(value) >= 16 && len(value)%4 == 0 {
		return "base64"
	}
	// JWT
	if matched, _ := regexp.MatchString(`^[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+$`, value); matched {
		return "jwt"
	}
	// Port
	if strings.ToLower(key) == "port" {
		if port, err := strconv.Atoi(value); err == nil && port >= 1 && port <= 65535 {
			return "port"
		}
	}
	return ""
}

func inferDescription(key string) string {
	// Convert KEY_NAME to "Key name"
	words := strings.Split(key, "_")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		}
	}
	return strings.Join(words, " ")
}

// ToSchema converts inferred variables to a Schema.
func (r *Result) ToSchema(version string) *schema.Schema {
	s := &schema.Schema{
		Version: version,
		Env:     r.Variables,
	}
	return s
}

// GenerateYAML produces a YAML string from the inferred schema.
func (r *Result) GenerateYAML(version string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "version: \"%s\"\n", version)
	b.WriteString("env:\n")

	// Sort keys for consistent output
	var keys []string
	for k := range r.Variables {
		keys = append(keys, k)
	}
	for _, k := range keys {
		v := r.Variables[k]
		fmt.Fprintf(&b, "  %s:\n", k)
		fmt.Fprintf(&b, "    type: %s\n", v.Type)
		if v.Description != "" {
			fmt.Fprintf(&b, "    description: \"%s\"\n", v.Description)
		}
		if v.Required {
			b.WriteString("    required: true\n")
		}
		if v.Default != nil {
			fmt.Fprintf(&b, "    default: %v\n", v.Default)
		}
		if v.Format != "" {
			fmt.Fprintf(&b, "    format: %s\n", v.Format)
		}
		if v.Sensitive {
			b.WriteString("    sensitive: true\n")
		}
		if v.Min != nil {
			fmt.Fprintf(&b, "    min: %v\n", v.Min)
		}
		if v.Max != nil {
			fmt.Fprintf(&b, "    max: %v\n", v.Max)
		}
		if v.Separator != "" {
			fmt.Fprintf(&b, "    separator: \"%s\"\n", v.Separator)
		}
	}

	return b.String()
}
