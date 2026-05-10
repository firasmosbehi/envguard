package validator

import (
	"testing"

	"github.com/envguard/envguard/internal/schema"
)

func ptr(b bool) *bool {
	return &b
}

func TestAllowEmpty(t *testing.T) {
	tests := []struct {
		name      string
		variable  *schema.Variable
		value     string
		wantValid bool
		wantRule  string
	}{
		{
			name:      "allowEmpty false with empty string",
			variable:  &schema.Variable{Type: schema.TypeString, AllowEmpty: ptr(false)},
			value:     "",
			wantValid: false,
			wantRule:  "allowEmpty",
		},
		{
			name:      "allowEmpty false with whitespace",
			variable:  &schema.Variable{Type: schema.TypeString, AllowEmpty: ptr(false)},
			value:     "   ",
			wantValid: false,
			wantRule:  "allowEmpty",
		},
		{
			name:      "allowEmpty false with value",
			variable:  &schema.Variable{Type: schema.TypeString, AllowEmpty: ptr(false)},
			value:     "hello",
			wantValid: true,
		},
		{
			name:      "allowEmpty true with empty string",
			variable:  &schema.Variable{Type: schema.TypeString, AllowEmpty: ptr(true)},
			value:     "",
			wantValid: true,
		},
		{
			name:      "default allowEmpty (nil) with empty string",
			variable:  &schema.Variable{Type: schema.TypeString},
			value:     "",
			wantValid: true,
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

func TestContains(t *testing.T) {
	tests := []struct {
		name      string
		variable  *schema.Variable
		value     string
		wantValid bool
		wantRule  string
	}{
		{
			name:      "contains found",
			variable:  &schema.Variable{Type: schema.TypeArray, Separator: ",", Contains: "admin"},
			value:     "read,write,admin",
			wantValid: true,
		},
		{
			name:      "contains not found",
			variable:  &schema.Variable{Type: schema.TypeArray, Separator: ",", Contains: "admin"},
			value:     "read,write",
			wantValid: false,
			wantRule:  "contains",
		},
		{
			name:      "contains with spaces",
			variable:  &schema.Variable{Type: schema.TypeArray, Separator: ",", Contains: "admin"},
			value:     " read , write , admin ",
			wantValid: true,
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

func TestDependsOn(t *testing.T) {
	tests := []struct {
		name      string
		schema    *schema.Schema
		envVars   map[string]string
		wantValid bool
		wantRule  string
	}{
		{
			name: "conditional required met",
			schema: &schema.Schema{
				Version: "1.0",
				Env: map[string]*schema.Variable{
					"HTTPS":     {Type: schema.TypeBoolean},
					"SSL_CERT":  {Type: schema.TypeString, DependsOn: "HTTPS", When: "true"},
				},
			},
			envVars:   map[string]string{"HTTPS": "true", "SSL_CERT": "/path/to/cert.pem"},
			wantValid: true,
		},
		{
			name: "conditional required not met - missing dependent value",
			schema: &schema.Schema{
				Version: "1.0",
				Env: map[string]*schema.Variable{
					"HTTPS":     {Type: schema.TypeBoolean},
					"SSL_CERT":  {Type: schema.TypeString, DependsOn: "HTTPS", When: "true"},
				},
			},
			envVars:   map[string]string{"HTTPS": "true", "SSL_CERT": ""},
			wantValid: false,
			wantRule:  "required",
		},
		{
			name: "conditional not triggered",
			schema: &schema.Schema{
				Version: "1.0",
				Env: map[string]*schema.Variable{
					"HTTPS":     {Type: schema.TypeBoolean},
					"SSL_CERT":  {Type: schema.TypeString, DependsOn: "HTTPS", When: "true"},
				},
			},
			envVars:   map[string]string{"HTTPS": "false"},
			wantValid: true,
		},
		{
			name: "conditional with missing dependency",
			schema: &schema.Schema{
				Version: "1.0",
				Env: map[string]*schema.Variable{
					"HTTPS":     {Type: schema.TypeBoolean},
					"SSL_CERT":  {Type: schema.TypeString, DependsOn: "HTTPS", When: "true"},
				},
			},
			envVars:   map[string]string{},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Validate(tt.schema, tt.envVars, false, "")
			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v; errors = %v", result.Valid, tt.wantValid, result.Errors)
			}
			if !tt.wantValid && len(result.Errors) > 0 && result.Errors[0].Rule != tt.wantRule {
				t.Errorf("Rule = %q, want %q", result.Errors[0].Rule, tt.wantRule)
			}
		})
	}
}
