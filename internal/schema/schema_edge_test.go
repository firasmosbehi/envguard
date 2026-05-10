package schema

import (
	"testing"
)

func TestValidateSchemaEmptyEnum(t *testing.T) {
	s := Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString, Enum: []any{}},
		},
	}
	err := s.Validate()
	if err == nil {
		t.Error("expected error for empty enum")
	}
}

func TestValidateSchemaNilEnum(t *testing.T) {
	// When enum is not specified, it should be nil and pass validation
	s := Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString},
		},
	}
	err := s.Validate()
	if err != nil {
		t.Errorf("unexpected error for nil enum: %v", err)
	}
}

func TestValidateSchemaFloatEnumWithIntValues(t *testing.T) {
	// YAML may parse integers as float64, but they should still be valid
	s := Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"MODE": {Type: TypeFloat, Enum: []any{float64(1), float64(2)}},
		},
	}
	err := s.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateSchemaIntegerDefaultZero(t *testing.T) {
	// Zero is a valid default for integer
	s := Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"COUNT": {Type: TypeInteger, Default: 0},
		},
	}
	err := s.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateSchemaBooleanDefaultFalse(t *testing.T) {
	// false is a valid default for boolean
	s := Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"DEBUG": {Type: TypeBoolean, Default: false},
		},
	}
	err := s.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateSchemaStringDefaultEmpty(t *testing.T) {
	// Empty string is a valid default
	s := Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"NAME": {Type: TypeString, Default: ""},
		},
	}
	err := s.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateSchemaPatternOnString(t *testing.T) {
	s := Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString, Pattern: "^[a-z]+$"},
		},
	}
	err := s.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateSchemaComplexPattern(t *testing.T) {
	s := Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"URL": {Type: TypeString, Pattern: `^https?://[\w.-]+\.[a-z]{2,}$`},
		},
	}
	err := s.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateSchemaUnicodeInDescription(t *testing.T) {
	s := Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString, Description: "日本語の説明"},
		},
	}
	err := s.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateSchemaVeryLongDescription(t *testing.T) {
	s := Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString, Description: string(make([]byte, 10000))},
		},
	}
	err := s.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateSchemaManyVariables(t *testing.T) {
	env := make(map[string]*Variable)
	for i := 0; i < 1000; i++ {
		env["VAR_0000"] = &Variable{Type: TypeString}
	}
	s := Schema{
		Version: "1.0",
		Env: env,
	}
	err := s.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestIsEnvVarNameValidEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{"A", true},
		{"_", true},
		{"A1", true},
		{"1A", false},
		{"A_B", true},
		{"A-B", false},
		{"A.B", false},
		{"A B", false},
		{"A\tB", false},
		{"A\nB", false},
		{"", false},
		{"________", true},
		{"A1B2C3", true},
	}

	for _, tt := range tests {
		got := IsEnvVarNameValid(tt.name)
		if got != tt.valid {
			t.Errorf("IsEnvVarNameValid(%q) = %v, want %v", tt.name, got, tt.valid)
		}
	}
}


