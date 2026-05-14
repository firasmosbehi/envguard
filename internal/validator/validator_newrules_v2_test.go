package validator

import (
	"testing"

	"github.com/envguard/envguard/internal/schema"
)

func TestPrefixSuffix(t *testing.T) {
	tests := []struct {
		name    string
		prefix  string
		suffix  string
		value   string
		isValid bool
	}{
		{"prefix match", "pre_", "", "pre_value", true},
		{"prefix mismatch", "pre_", "", "value", false},
		{"suffix match", "", "_suf", "value_suf", true},
		{"suffix mismatch", "", "_suf", "value", false},
		{"both match", "pre_", "_suf", "pre_value_suf", true},
		{"both mismatch", "pre_", "_suf", "value", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &schema.Schema{
				Version: "1.0",
				Env: map[string]*schema.Variable{
					"VAR": {
						Type:   schema.TypeString,
						Prefix: tt.prefix,
						Suffix: tt.suffix,
					},
				},
			}
			envVars := map[string]string{"VAR": tt.value}
			result := Validate(s, envVars, false, "")
			if result.Valid != tt.isValid {
				t.Errorf("expected valid=%v, got %v (errors: %v)", tt.isValid, result.Valid, result.Errors)
			}
		})
	}
}

func TestMultipleOf(t *testing.T) {
	tests := []struct {
		name       string
		typ        schema.Type
		multipleOf any
		value      string
		isValid    bool
	}{
		{"integer multipleOf 5 valid", schema.TypeInteger, 5, "15", true},
		{"integer multipleOf 5 invalid", schema.TypeInteger, 5, "13", false},
		{"integer multipleOf 2 valid", schema.TypeInteger, 2, "0", true},
		{"float multipleOf 0.5 valid", schema.TypeFloat, 0.5, "1.5", true},
		{"float multipleOf 0.5 invalid", schema.TypeFloat, 0.5, "1.3", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &schema.Schema{
				Version: "1.0",
				Env: map[string]*schema.Variable{
					"VAR": {
						Type:       tt.typ,
						MultipleOf: tt.multipleOf,
					},
				},
			}
			envVars := map[string]string{"VAR": tt.value}
			result := Validate(s, envVars, false, "")
			if result.Valid != tt.isValid {
				t.Errorf("expected valid=%v, got %v (errors: %v)", tt.isValid, result.Valid, result.Errors)
			}
		})
	}
}

func TestArrayUniqueItems(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"TAGS": {
				Type:        schema.TypeArray,
				Separator:   ",",
				UniqueItems: true,
			},
		},
	}

	tests := []struct {
		name    string
		value   string
		isValid bool
	}{
		{"unique items", "a,b,c", true},
		{"duplicate items", "a,b,a", false},
		{"single item", "a", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envVars := map[string]string{"TAGS": tt.value}
			result := Validate(s, envVars, false, "")
			if result.Valid != tt.isValid {
				t.Errorf("expected valid=%v, got %v (errors: %v)", tt.isValid, result.Valid, result.Errors)
			}
		})
	}
}

func TestArrayItemPattern(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"EMAILS": {
				Type:        schema.TypeArray,
				Separator:   ",",
				ItemPattern: `^[^@]+@[^@]+\.[^@]+$`,
			},
		},
	}

	tests := []struct {
		name    string
		value   string
		isValid bool
	}{
		{"valid emails", "a@b.com,c@d.com", true},
		{"invalid email", "a@b.com,invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envVars := map[string]string{"EMAILS": tt.value}
			result := Validate(s, envVars, false, "")
			if result.Valid != tt.isValid {
				t.Errorf("expected valid=%v, got %v (errors: %v)", tt.isValid, result.Valid, result.Errors)
			}
		})
	}
}

func TestArrayItemType(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"PORTS": {
				Type:      schema.TypeArray,
				Separator: ",",
				ItemType:  schema.TypeInteger,
			},
		},
	}

	tests := []struct {
		name    string
		value   string
		isValid bool
	}{
		{"valid integers", "80,443,8080", true},
		{"invalid integer", "80,abc,8080", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envVars := map[string]string{"PORTS": tt.value}
			result := Validate(s, envVars, false, "")
			if result.Valid != tt.isValid {
				t.Errorf("expected valid=%v, got %v (errors: %v)", tt.isValid, result.Valid, result.Errors)
			}
		})
	}
}

func TestArrayNotEmpty(t *testing.T) {
	notEmpty := true
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"ITEMS": {
				Type:      schema.TypeArray,
				Separator: ",",
				NotEmpty:  &notEmpty,
			},
		},
	}

	tests := []struct {
		name    string
		value   string
		isValid bool
	}{
		{"non-empty", "a,b", true},
		{"truly empty", "", false},
		{"single item", "a", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envVars := map[string]string{"ITEMS": tt.value}
			result := Validate(s, envVars, false, "")
			if result.Valid != tt.isValid {
				t.Errorf("expected valid=%v, got %v (errors: %v)", tt.isValid, result.Valid, result.Errors)
			}
		})
	}
}
