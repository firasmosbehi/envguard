package validator

import (
	"testing"

	"github.com/envguard/envguard/internal/schema"
)

func TestValidateArray(t *testing.T) {
	min2 := 2
	max5 := 5

	tests := []struct {
		name      string
		variable  *schema.Variable
		value     string
		wantValid bool
		wantRule  string
	}{
		{
			name:      "valid csv array",
			variable:  &schema.Variable{Type: schema.TypeArray, Separator: ","},
			value:     "a,b,c",
			wantValid: true,
		},
		{
			name:      "valid pipe-separated array",
			variable:  &schema.Variable{Type: schema.TypeArray, Separator: "|"},
			value:     "x|y|z",
			wantValid: true,
		},
		{
			name:      "empty value required",
			variable:  &schema.Variable{Type: schema.TypeArray, Separator: ",", Required: true},
			value:     "",
			wantValid: false,
			wantRule:  "required",
		},
		{
			name:      "minLength satisfied",
			variable:  &schema.Variable{Type: schema.TypeArray, Separator: ",", MinLength: &min2},
			value:     "a,b,c",
			wantValid: true,
		},
		{
			name:      "minLength not satisfied",
			variable:  &schema.Variable{Type: schema.TypeArray, Separator: ",", MinLength: &min2},
			value:     "a",
			wantValid: false,
			wantRule:  "minLength",
		},
		{
			name:      "maxLength satisfied",
			variable:  &schema.Variable{Type: schema.TypeArray, Separator: ",", MaxLength: &max5},
			value:     "a,b,c",
			wantValid: true,
		},
		{
			name:      "maxLength exceeded",
			variable:  &schema.Variable{Type: schema.TypeArray, Separator: ",", MaxLength: &max5},
			value:     "a,b,c,d,e,f",
			wantValid: false,
			wantRule:  "maxLength",
		},
		{
			name:      "enum all match",
			variable:  &schema.Variable{Type: schema.TypeArray, Separator: ",", Enum: []any{"read", "write", "admin"}},
			value:     "read,write",
			wantValid: true,
		},
		{
			name:      "enum one mismatch",
			variable:  &schema.Variable{Type: schema.TypeArray, Separator: ",", Enum: []any{"read", "write"}},
			value:     "read,delete",
			wantValid: false,
			wantRule:  "enum",
		},
		{
			name:      "single item array",
			variable:  &schema.Variable{Type: schema.TypeArray, Separator: ","},
			value:     "only-one",
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
