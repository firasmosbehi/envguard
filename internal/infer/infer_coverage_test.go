package infer

import (
	"strings"
	"testing"

	"github.com/envguard/envguard/internal/schema"
)

func TestInferVariableFloatWholeNumber(t *testing.T) {
	// A float that is a whole number should still be inferred as float
	v := inferVariable("RATIO", "3.0")
	if v.Type != schema.TypeFloat {
		t.Errorf("expected float type, got %q", v.Type)
	}
}

func TestInferVariableSingleItemArray(t *testing.T) {
	// Single item with comma shouldn't be array
	v := inferVariable("LIST", "a")
	if v.Type == schema.TypeArray {
		t.Error("expected single item not to be inferred as array")
	}
}

func TestInferVariableArrayWithURL(t *testing.T) {
	// Array containing URL-like value with //
	v := inferVariable("URLS", "http://a.com,http://b.com")
	if v.Type == schema.TypeArray {
		t.Error("expected URLs with // not to be inferred as array")
	}
}

func TestInferVariableFormatDetection(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{"550e8400-e29b-41d4-a716-446655440000", "uuid"},
		{"admin@example.com", "email"},
		{"https://example.com", "url"},
		{"192.168.1.1", "ip"},
		{"aabbccdd11223344", "hex"},
		{"v1.2.3", "semver"},
		{"30s", "duration"},
		{"dGVzdHRlc3R0ZXN0dGVzdA==", "base64"},
		{"eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.sig", "jwt"},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			v := inferVariable("TEST", tt.value)
			if v.Format != tt.want {
				t.Errorf("format = %q, want %q", v.Format, tt.want)
			}
		})
	}
}

func TestInferVariablePortFormat(t *testing.T) {
	// PORT with integer value is inferred as TypeInteger before format detection
	v := inferVariable("PORT", "8080")
	if v.Type != schema.TypeInteger {
		t.Errorf("expected integer type, got %q", v.Type)
	}
	if v.Min != 1 || v.Max != 65535 {
		t.Errorf("expected port min=1 max=65535, got min=%v max=%v", v.Min, v.Max)
	}
}

func TestInferVariableHostnameFormat(t *testing.T) {
	v := inferVariable("HOST", "example.com")
	if v.Format != "hostname" {
		t.Errorf("expected hostname format, got %q", v.Format)
	}
}

func TestInferVariableEmailFormatFromKey(t *testing.T) {
	v := inferVariable("SMTP_EMAIL", "test@example.com")
	if v.Format != "email" {
		t.Errorf("expected email format, got %q", v.Format)
	}
}

func TestInferVariableURLFormatFromKey(t *testing.T) {
	v := inferVariable("API_ENDPOINT", "https://api.example.com")
	if v.Format != "url" {
		t.Errorf("expected url format, got %q", v.Format)
	}
}

func TestInferVariableSensitiveFromKey(t *testing.T) {
	sensitiveKeys := []string{"PASSWORD", "SECRET_KEY", "AUTH_TOKEN", "API_SECRET"}
	for _, key := range sensitiveKeys {
		t.Run(key, func(t *testing.T) {
			v := inferVariable(key, "some-value")
			if !v.Sensitive {
				t.Errorf("expected %s to be sensitive", key)
			}
		})
	}
}

func TestInferVariableBooleanVariations(t *testing.T) {
	boolValues := []string{"true", "false", "1", "0", "yes", "no", "on", "off", "TRUE", "FALSE"}
	for _, val := range boolValues {
		t.Run(val, func(t *testing.T) {
			v := inferVariable("FLAG", val)
			if v.Type != schema.TypeBoolean {
				t.Errorf("expected boolean for %q, got %q", val, v.Type)
			}
		})
	}
}

func TestToSchema(t *testing.T) {
	r := FromEnv(map[string]string{
		"DEBUG": "true",
		"PORT":  "3000",
	})

	s := r.ToSchema("1.0")
	if s.Version != "1.0" {
		t.Errorf("expected version 1.0, got %q", s.Version)
	}
	if len(s.Env) != 2 {
		t.Errorf("expected 2 variables, got %d", len(s.Env))
	}
	if s.Env["DEBUG"].Type != schema.TypeBoolean {
		t.Errorf("expected DEBUG to be boolean, got %q", s.Env["DEBUG"].Type)
	}
}

func TestGenerateYAMLWithAllFields(t *testing.T) {
	r := FromEnv(map[string]string{
		"DEBUG":    "true",
		"PORT":     "3000",
		"DATABASE": "postgresql://localhost",
		"API_KEY":  "secret",
	})

	// Manually adjust some fields to test YAML generation
	r.Variables["DEBUG"].Required = true
	r.Variables["PORT"].Min = 1
	r.Variables["PORT"].Max = 65535
	r.Variables["DATABASE"].Format = "url"
	r.Variables["API_KEY"].Sensitive = true

	yaml := r.GenerateYAML("1.0")

	checks := []string{
		`version: "1.0"`,
		"DEBUG:",
		"PORT:",
		"DATABASE:",
		"API_KEY:",
		"required: true",
		"min:",
		"max:",
		"format:",
		"sensitive: true",
	}

	for _, check := range checks {
		if !strings.Contains(yaml, check) {
			t.Errorf("expected YAML to contain %q", check)
		}
	}
}

func TestGenerateYAMLSeparator(t *testing.T) {
	r := FromEnv(map[string]string{
		"TAGS": "a,b,c",
	})

	yaml := r.GenerateYAML("1.0")
	if !strings.Contains(yaml, `separator: ","`) {
		t.Errorf("expected separator in YAML, got:\n%s", yaml)
	}
}
