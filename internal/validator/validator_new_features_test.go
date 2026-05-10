package validator

import (
	"testing"

	"github.com/envguard/envguard/internal/schema"
)

func TestValidateMinMax(t *testing.T) {
	tests := []struct {
		name      string
		variable  *schema.Variable
		value     string
		wantValid bool
		wantRule  string
	}{
		{
			name:      "integer within range",
			variable:  &schema.Variable{Type: schema.TypeInteger, Min: 1024, Max: 65535},
			value:     "8080",
			wantValid: true,
		},
		{
			name:      "integer below min",
			variable:  &schema.Variable{Type: schema.TypeInteger, Min: 1024},
			value:     "80",
			wantValid: false,
			wantRule:  "min",
		},
		{
			name:      "integer above max",
			variable:  &schema.Variable{Type: schema.TypeInteger, Max: 65535},
			value:     "70000",
			wantValid: false,
			wantRule:  "max",
		},
		{
			name:      "float within range",
			variable:  &schema.Variable{Type: schema.TypeFloat, Min: 0.0, Max: 1.0},
			value:     "0.5",
			wantValid: true,
		},
		{
			name:      "float below min",
			variable:  &schema.Variable{Type: schema.TypeFloat, Min: 0.0},
			value:     "-0.1",
			wantValid: false,
			wantRule:  "min",
		},
		{
			name:      "float above max",
			variable:  &schema.Variable{Type: schema.TypeFloat, Max: 1.0},
			value:     "1.5",
			wantValid: false,
			wantRule:  "max",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &schema.Schema{
				Version: "1.0",
				Env:     map[string]*schema.Variable{"TEST": tt.variable},
			}
			result := Validate(s, map[string]string{"TEST": tt.value}, false, "")
			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v; errors = %v", result.Valid, tt.wantValid, result.Errors)
			}
			if !tt.wantValid && len(result.Errors) > 0 && result.Errors[0].Rule != tt.wantRule {
				t.Errorf("Rule = %q, want %q", result.Errors[0].Rule, tt.wantRule)
			}
		})
	}
}

func TestValidateStringLength(t *testing.T) {
	min5 := 5
	max10 := 10

	tests := []struct {
		name      string
		variable  *schema.Variable
		value     string
		wantValid bool
		wantRule  string
	}{
		{
			name:      "string within length range",
			variable:  &schema.Variable{Type: schema.TypeString, MinLength: &min5, MaxLength: &max10},
			value:     "hello",
			wantValid: true,
		},
		{
			name:      "string too short",
			variable:  &schema.Variable{Type: schema.TypeString, MinLength: &min5},
			value:     "hi",
			wantValid: false,
			wantRule:  "minLength",
		},
		{
			name:      "string too long",
			variable:  &schema.Variable{Type: schema.TypeString, MaxLength: &max10},
			value:     "hello world!",
			wantValid: false,
			wantRule:  "maxLength",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &schema.Schema{
				Version: "1.0",
				Env:     map[string]*schema.Variable{"TEST": tt.variable},
			}
			result := Validate(s, map[string]string{"TEST": tt.value}, false, "")
			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v; errors = %v", result.Valid, tt.wantValid, result.Errors)
			}
			if !tt.wantValid && len(result.Errors) > 0 && result.Errors[0].Rule != tt.wantRule {
				t.Errorf("Rule = %q, want %q", result.Errors[0].Rule, tt.wantRule)
			}
		})
	}
}

func TestValidateFormat(t *testing.T) {
	tests := []struct {
		name      string
		variable  *schema.Variable
		value     string
		wantValid bool
		wantRule  string
	}{
		{
			name:      "valid email",
			variable:  &schema.Variable{Type: schema.TypeString, Format: "email"},
			value:     "user@example.com",
			wantValid: true,
		},
		{
			name:      "invalid email",
			variable:  &schema.Variable{Type: schema.TypeString, Format: "email"},
			value:     "not-an-email",
			wantValid: false,
			wantRule:  "format",
		},
		{
			name:      "valid url",
			variable:  &schema.Variable{Type: schema.TypeString, Format: "url"},
			value:     "https://example.com/path",
			wantValid: true,
		},
		{
			name:      "invalid url - no scheme",
			variable:  &schema.Variable{Type: schema.TypeString, Format: "url"},
			value:     "example.com",
			wantValid: false,
			wantRule:  "format",
		},
		{
			name:      "valid uuid",
			variable:  &schema.Variable{Type: schema.TypeString, Format: "uuid"},
			value:     "550e8400-e29b-41d4-a716-446655440000",
			wantValid: true,
		},
		{
			name:      "invalid uuid",
			variable:  &schema.Variable{Type: schema.TypeString, Format: "uuid"},
			value:     "not-a-uuid",
			wantValid: false,
			wantRule:  "format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &schema.Schema{
				Version: "1.0",
				Env:     map[string]*schema.Variable{"TEST": tt.variable},
			}
			result := Validate(s, map[string]string{"TEST": tt.value}, false, "")
			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v; errors = %v", result.Valid, tt.wantValid, result.Errors)
			}
			if !tt.wantValid && len(result.Errors) > 0 && result.Errors[0].Rule != tt.wantRule {
				t.Errorf("Rule = %q, want %q", result.Errors[0].Rule, tt.wantRule)
			}
		})
	}
}

func TestValidateDisallow(t *testing.T) {
	tests := []struct {
		name      string
		variable  *schema.Variable
		value     string
		wantValid bool
	}{
		{
			name:      "allowed value",
			variable:  &schema.Variable{Type: schema.TypeString, Disallow: []string{"undefined", "null"}},
			value:     "my-api-key",
			wantValid: true,
		},
		{
			name:      "disallowed value undefined",
			variable:  &schema.Variable{Type: schema.TypeString, Disallow: []string{"undefined", "null"}},
			value:     "undefined",
			wantValid: false,
		},
		{
			name:      "disallowed value null",
			variable:  &schema.Variable{Type: schema.TypeString, Disallow: []string{"undefined", "null"}},
			value:     "null",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &schema.Schema{
				Version: "1.0",
				Env:     map[string]*schema.Variable{"TEST": tt.variable},
			}
			result := Validate(s, map[string]string{"TEST": tt.value}, false, "")
			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v; errors = %v", result.Valid, tt.wantValid, result.Errors)
			}
		})
	}
}

func TestValidateEnvironmentRules(t *testing.T) {
	tests := []struct {
		name      string
		variable  *schema.Variable
		envName   string
		value     string
		wantValid bool
		wantRule  string
	}{
		{
			name:      "requiredIn - required in production, missing",
			variable:  &schema.Variable{Type: schema.TypeString, RequiredIn: []string{"production"}},
			envName:   "production",
			value:     "",
			wantValid: false,
			wantRule:  "required",
		},
		{
			name:      "requiredIn - required in production, present",
			variable:  &schema.Variable{Type: schema.TypeString, RequiredIn: []string{"production"}},
			envName:   "production",
			value:     "ok",
			wantValid: true,
		},
		{
			name:      "requiredIn - not required in development, missing",
			variable:  &schema.Variable{Type: schema.TypeString, RequiredIn: []string{"production"}},
			envName:   "development",
			value:     "",
			wantValid: true,
		},
		{
			name:      "requiredIn - no env specified, not required",
			variable:  &schema.Variable{Type: schema.TypeString, RequiredIn: []string{"production"}},
			envName:   "",
			value:     "",
			wantValid: true,
		},
		{
			name:      "devOnly - ignored in production",
			variable:  &schema.Variable{Type: schema.TypeString, DevOnly: true},
			envName:   "production",
			value:     "",
			wantValid: true,
		},
		{
			name:      "devOnly - required in development, missing",
			variable:  &schema.Variable{Type: schema.TypeString, DevOnly: true},
			envName:   "development",
			value:     "",
			wantValid: false,
			wantRule:  "required",
		},
		{
			name:      "devOnly - required in dev, missing",
			variable:  &schema.Variable{Type: schema.TypeString, DevOnly: true},
			envName:   "dev",
			value:     "",
			wantValid: false,
			wantRule:  "required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &schema.Schema{
				Version: "1.0",
				Env:     map[string]*schema.Variable{"TEST": tt.variable},
			}
			result := Validate(s, map[string]string{"TEST": tt.value}, false, tt.envName)
			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v; errors = %v", result.Valid, tt.wantValid, result.Errors)
			}
			if !tt.wantValid && len(result.Errors) > 0 && result.Errors[0].Rule != tt.wantRule {
				t.Errorf("Rule = %q, want %q", result.Errors[0].Rule, tt.wantRule)
			}
		})
	}
}
