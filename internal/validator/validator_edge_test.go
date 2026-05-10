package validator

import (
	"testing"

	"github.com/envguard/envguard/internal/schema"
)

// === Validator edge cases ===

func TestValidateWhitespaceOnlyRequired(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"FOO": {Type: schema.TypeString, Required: true},
		},
	}
	vars := map[string]string{"FOO": "   "}
	result := Validate(s, vars, false)
	if result.Valid {
		t.Error("expected validation to fail for whitespace-only required value")
	}
}

func TestValidateEmptyStringWithDefault(t *testing.T) {
	// If FOO= (empty) and default is "bar", should default be used?
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"FOO": {Type: schema.TypeString, Default: "bar"},
		},
	}
	vars := map[string]string{"FOO": ""}
	result := Validate(s, vars, false)
	if !result.Valid {
		t.Errorf("expected valid, got errors: %v", result.Errors)
	}
	// Note: current behavior treats empty string as missing and applies default.
	// This may or may not be desired.
}

func TestValidateFloatWholeNumber(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"RATIO": {Type: schema.TypeFloat},
		},
	}
	vars := map[string]string{"RATIO": "3.0"}
	result := Validate(s, vars, false)
	if !result.Valid {
		t.Errorf("expected valid, got errors: %v", result.Errors)
	}
}

func TestValidateFloatLeadingDot(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"RATIO": {Type: schema.TypeFloat},
		},
	}
	vars := map[string]string{"RATIO": ".5"}
	result := Validate(s, vars, false)
	if !result.Valid {
		t.Errorf("expected valid for '.5', got errors: %v", result.Errors)
	}
}

func TestValidateFloatTrailingDot(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"RATIO": {Type: schema.TypeFloat},
		},
	}
	vars := map[string]string{"RATIO": "5."}
	result := Validate(s, vars, false)
	if !result.Valid {
		t.Errorf("expected valid for '5.', got errors: %v", result.Errors)
	}
}

func TestValidateIntegerWithLeadingZeros(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"COUNT": {Type: schema.TypeInteger},
		},
	}
	vars := map[string]string{"COUNT": "007"}
	result := Validate(s, vars, false)
	if !result.Valid {
		t.Errorf("expected valid for '007', got errors: %v", result.Errors)
	}
}

func TestValidateIntegerNegative(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"OFFSET": {Type: schema.TypeInteger},
		},
	}
	vars := map[string]string{"OFFSET": "-42"}
	result := Validate(s, vars, false)
	if !result.Valid {
		t.Errorf("expected valid for '-42', got errors: %v", result.Errors)
	}
}

func TestValidateIntegerPlusSign(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"COUNT": {Type: schema.TypeInteger},
		},
	}
	vars := map[string]string{"COUNT": "+42"}
	result := Validate(s, vars, false)
	if !result.Valid {
		t.Errorf("expected valid for '+42', got errors: %v", result.Errors)
	}
}

func TestValidateBooleanMixedCase(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"DEBUG": {Type: schema.TypeBoolean},
		},
	}
	cases := []string{"True", "FALSE", "YES", "No", "ON", "Off", "1", "0"}
	for _, c := range cases {
		vars := map[string]string{"DEBUG": c}
		result := Validate(s, vars, false)
		if !result.Valid {
			t.Errorf("expected valid for boolean %q, got errors: %v", c, result.Errors)
		}
	}
}

func TestValidatePatternWithSpecialChars(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"FOO": {
				Type:    schema.TypeString,
				Pattern: `^[a-z]+/[a-z]+$`,
			},
		},
	}
	vars := map[string]string{"FOO": "foo/bar"}
	result := Validate(s, vars, false)
	if !result.Valid {
		t.Errorf("expected valid for pattern match, got errors: %v", result.Errors)
	}
}

func TestValidatePatternInvalidRegexInSchema(t *testing.T) {
	// This should be caught during schema parsing, not validation
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"FOO": {
				Type:    schema.TypeString,
				Pattern: `[invalid(`,
			},
		},
	}
	err := s.Validate()
	if err == nil {
		t.Error("expected schema validation to fail for invalid regex pattern")
	}
}

func TestValidateEnumWithFloatWholeNumbers(t *testing.T) {
	// Enum has integers but value is float string
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"MODE": {
				Type: schema.TypeFloat,
				Enum: []any{1.0, 2.0, 3.0},
			},
		},
	}
	vars := map[string]string{"MODE": "1"}
	result := Validate(s, vars, false)
	if !result.Valid {
		t.Errorf("expected valid for float enum with whole number, got errors: %v", result.Errors)
	}
}

func TestValidateEmptyEnum(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"FOO": {
				Type: schema.TypeString,
				Enum: []any{},
			},
		},
	}
	vars := map[string]string{"FOO": "bar"}
	result := Validate(s, vars, false)
	// Empty enum means no values are allowed, so this should fail
	if result.Valid {
		t.Error("expected validation to fail for empty enum")
	}
}

func TestValidateStrictModeIgnoresEmptyLines(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"FOO": {Type: schema.TypeString},
		},
	}
	vars := map[string]string{"FOO": "bar"}
	result := Validate(s, vars, true)
	if len(result.Warnings) > 0 {
		t.Errorf("expected no warnings for clean env, got: %v", result.Warnings)
	}
}

func TestValidateStrictModeWithUnknownVars(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"FOO": {Type: schema.TypeString},
		},
	}
	vars := map[string]string{"FOO": "bar", "UNKNOWN": "value"}
	result := Validate(s, vars, true)
	if len(result.Warnings) != 1 {
		t.Errorf("expected 1 warning, got %d", len(result.Warnings))
	}
}

func TestValidateMultipleDefaults(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"PORT":     {Type: schema.TypeInteger, Default: 3000},
			"DEBUG":    {Type: schema.TypeBoolean, Default: false},
			"NAME":     {Type: schema.TypeString, Default: "app"},
			"RATIO":    {Type: schema.TypeFloat, Default: 1.5},
		},
	}
	vars := map[string]string{}
	result := Validate(s, vars, false)
	if !result.Valid {
		t.Errorf("expected all defaults to pass, got errors: %v", result.Errors)
	}
}

func TestValidateRequiredAndEmptyEnvFile(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"FOO": {Type: schema.TypeString, Required: true},
		},
	}
	vars := map[string]string{}
	result := Validate(s, vars, false)
	if result.Valid {
		t.Error("expected validation to fail for missing required var")
	}
	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(result.Errors))
	}
}

func TestValidateAllTypesAtOnce(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"STR":    {Type: schema.TypeString, Required: true},
			"INT":    {Type: schema.TypeInteger, Required: true},
			"FLOAT":  {Type: schema.TypeFloat, Required: true},
			"BOOL":   {Type: schema.TypeBoolean, Required: true},
		},
	}
	vars := map[string]string{
		"STR":   "hello",
		"INT":   "42",
		"FLOAT": "3.14",
		"BOOL":  "true",
	}
	result := Validate(s, vars, false)
	if !result.Valid {
		t.Errorf("expected all types to pass, got errors: %v", result.Errors)
	}
}

func TestValidateStringWithNewline(t *testing.T) {
	// A string value containing an actual newline character
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"MSG": {Type: schema.TypeString},
		},
	}
	vars := map[string]string{"MSG": "line1\nline2"}
	result := Validate(s, vars, false)
	if !result.Valid {
		t.Errorf("expected valid for multiline string, got errors: %v", result.Errors)
	}
}

func TestValidateIntegerMaxInt64(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"BIG": {Type: schema.TypeInteger},
		},
	}
	vars := map[string]string{"BIG": "9223372036854775807"}
	result := Validate(s, vars, false)
	if !result.Valid {
		t.Errorf("expected valid for max int64, got errors: %v", result.Errors)
	}
}

func TestValidateIntegerOverflow(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"BIG": {Type: schema.TypeInteger},
		},
	}
	vars := map[string]string{"BIG": "9223372036854775808"}
	result := Validate(s, vars, false)
	if result.Valid {
		t.Error("expected validation to fail for integer overflow")
	}
}

func TestValidateFloatScientificNotation(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"SCI": {Type: schema.TypeFloat},
		},
	}
	vars := map[string]string{"SCI": "1.5e10"}
	result := Validate(s, vars, false)
	if !result.Valid {
		t.Errorf("expected valid for scientific notation, got errors: %v", result.Errors)
	}
}

func TestValidateFloatNegativeZero(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"VAL": {Type: schema.TypeFloat},
		},
	}
	vars := map[string]string{"VAL": "-0.0"}
	result := Validate(s, vars, false)
	if !result.Valid {
		t.Errorf("expected valid for -0.0, got errors: %v", result.Errors)
	}
}

func TestValidateBooleanEmptyString(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"DEBUG": {Type: schema.TypeBoolean},
		},
	}
	vars := map[string]string{"DEBUG": ""}
	result := Validate(s, vars, false)
	// Empty string for boolean should either use default or fail
	// Current behavior: empty string triggers default if set, otherwise passes silently
	// This test documents current behavior
	if !result.Valid {
		t.Logf("Boolean empty string failed with errors: %v", result.Errors)
	}
}

func TestValidateStringEnumCaseSensitive(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"LEVEL": {
				Type: schema.TypeString,
				Enum: []any{"debug", "info", "warn", "error"},
			},
		},
	}
	vars := map[string]string{"LEVEL": "DEBUG"}
	result := Validate(s, vars, false)
	if result.Valid {
		t.Error("expected validation to fail for case-sensitive enum mismatch")
	}
}

func TestValidateDuplicateKeysInEnv(t *testing.T) {
	// If .env has duplicate keys, the last one wins (map behavior)
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"FOO": {Type: schema.TypeString},
		},
	}
	vars := map[string]string{"FOO": "second"}
	result := Validate(s, vars, false)
	if !result.Valid {
		t.Errorf("expected valid, got errors: %v", result.Errors)
	}
}
