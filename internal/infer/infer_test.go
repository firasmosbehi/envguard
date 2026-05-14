package infer

import (
	"testing"

	"github.com/envguard/envguard/internal/schema"
)

func TestInferVariable(t *testing.T) {
	tests := []struct {
		key      string
		value    string
		wantType schema.Type
		wantFmt  string
	}{
		{"DEBUG", "true", schema.TypeBoolean, ""},
		{"ENABLED", "1", schema.TypeBoolean, ""},
		{"PORT", "3000", schema.TypeInteger, ""},
		{"TIMEOUT", "30", schema.TypeInteger, ""},
		{"RATIO", "0.5", schema.TypeFloat, ""},
		{"API_KEY", "sk-abc123", schema.TypeString, ""},
		{"DATABASE_URL", "postgresql://localhost:5432/db", schema.TypeString, "url"},
		{"EMAIL", "admin@example.com", schema.TypeString, "email"},
		{"UUID", "550e8400-e29b-41d4-a716-446655440000", schema.TypeString, "uuid"},
		{"IP", "192.168.1.1", schema.TypeString, "ip"},
		{"VERSION", "1.2.3", schema.TypeString, "semver"},
		{"DURATION", "30s", schema.TypeString, "duration"},
		{"ARRAY", "a,b,c", schema.TypeArray, ""},
		{"JWT", "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U", schema.TypeString, "jwt"},
		{"HEX", "aabbccdd", schema.TypeString, "hex"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			v := inferVariable(tt.key, tt.value)
			if v.Type != tt.wantType {
				t.Errorf("type = %q, want %q", v.Type, tt.wantType)
			}
			if v.Format != tt.wantFmt {
				t.Errorf("format = %q, want %q", v.Format, tt.wantFmt)
			}
		})
	}
}

func TestInferSensitive(t *testing.T) {
	v := inferVariable("API_SECRET", "shhhh")
	if !v.Sensitive {
		t.Error("expected API_SECRET to be marked sensitive")
	}

	v2 := inferVariable("DB_PASSWORD", "secret")
	if !v2.Sensitive {
		t.Error("expected DB_PASSWORD to be marked sensitive")
	}
}

func TestInferDescription(t *testing.T) {
	if got := inferDescription("API_KEY"); got != "Api Key" {
		t.Errorf("inferDescription(API_KEY) = %q, want Api Key", got)
	}
	if got := inferDescription("DATABASE_URL"); got != "Database Url" {
		t.Errorf("inferDescription(DATABASE_URL) = %q, want Database Url", got)
	}
}

func TestGenerateYAML(t *testing.T) {
	r := FromEnv(map[string]string{
		"DEBUG": "true",
		"PORT":  "3000",
	})
	yaml := r.GenerateYAML("1.0")
	if yaml == "" {
		t.Error("expected non-empty YAML")
	}
	if !contains(yaml, "version: \"1.0\"") {
		t.Error("expected version in YAML")
	}
	if !contains(yaml, "DEBUG:") {
		t.Error("expected DEBUG in YAML")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
