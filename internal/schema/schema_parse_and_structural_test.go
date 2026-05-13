package schema

import (
	"os"
	"path/filepath"
	"testing"
)

// === ParseLenient tests ===

func TestParseLenientValidSchema(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")

	content := `
version: "1.0"
env:
  FOO:
    type: string
`
	os.WriteFile(schemaPath, []byte(content), 0644)

	s, err := ParseLenient(schemaPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Version != "1.0" {
		t.Errorf("version = %q, want 1.0", s.Version)
	}
	if _, ok := s.Env["FOO"]; !ok {
		t.Error("FOO should be in schema")
	}
}

func TestParseLenientMissingFile(t *testing.T) {
	_, err := ParseLenient("/nonexistent/schema.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestParseLenientInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := filepath.Join(tmpDir, "envguard.yaml")

	os.WriteFile(schemaPath, []byte(`not: [ valid yaml :::`), 0644)

	_, err := ParseLenient(schemaPath)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestParseLenientWithExtends(t *testing.T) {
	tmpDir := t.TempDir()
	basePath := filepath.Join(tmpDir, "base.yaml")
	childPath := filepath.Join(tmpDir, "child.yaml")

	baseContent := `
version: "1.0"
env:
  BASE_VAR:
    type: string
`
	childContent := `
version: "1.0"
extends: ./base.yaml
env:
  CHILD_VAR:
    type: integer
`
	os.WriteFile(basePath, []byte(baseContent), 0644)
	os.WriteFile(childPath, []byte(childContent), 0644)

	s, err := ParseLenient(childPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := s.Env["BASE_VAR"]; !ok {
		t.Error("BASE_VAR should be inherited")
	}
	if _, ok := s.Env["CHILD_VAR"]; !ok {
		t.Error("CHILD_VAR should be in child")
	}
}

func TestParseLenientCircularExtends(t *testing.T) {
	// ParseLenient does not detect circular extends (by design, it's lenient)
	// This test documents that behavior; Parse() does detect circular extends.
	t.Skip("ParseLenient does not implement circular extends detection")
}

// === validateEnumValue tests ===

func TestValidateEnumValueString(t *testing.T) {
	if err := validateEnumValue(TypeString, "valid"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := validateEnumValue(TypeString, 123); err == nil {
		t.Error("expected error for non-string enum value")
	}
}

func TestValidateEnumValueInteger(t *testing.T) {
	validInts := []any{int(1), int8(1), int16(1), int32(1), int64(1), uint(1), uint8(1), uint16(1), uint32(1), uint64(1), float64(1.0)}
	for _, v := range validInts {
		if err := validateEnumValue(TypeInteger, v); err != nil {
			t.Errorf("unexpected error for %T(%v): %v", v, v, err)
		}
	}

	if err := validateEnumValue(TypeInteger, float64(1.5)); err == nil {
		t.Error("expected error for non-integer float enum value")
	}
	if err := validateEnumValue(TypeInteger, "1"); err == nil {
		t.Error("expected error for string enum value with integer type")
	}
}

func TestValidateEnumValueFloat(t *testing.T) {
	validFloats := []any{int(1), int8(1), int16(1), int32(1), int64(1), uint(1), float32(1.5), float64(1.5)}
	for _, v := range validFloats {
		if err := validateEnumValue(TypeFloat, v); err != nil {
			t.Errorf("unexpected error for %T(%v): %v", v, v, err)
		}
	}

	if err := validateEnumValue(TypeFloat, "1.5"); err == nil {
		t.Error("expected error for string enum value with float type")
	}
}

func TestValidateEnumValueArray(t *testing.T) {
	if err := validateEnumValue(TypeArray, "valid"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := validateEnumValue(TypeArray, 123); err == nil {
		t.Error("expected error for non-string enum value with array type")
	}
}

func TestValidateEnumValueUnsupportedType(t *testing.T) {
	if err := validateEnumValue(TypeBoolean, true); err == nil {
		t.Error("expected error for boolean enum")
	}
}

// === toFloat64 tests ===

func TestToFloat64AllTypes(t *testing.T) {
	tests := []struct {
		input  any
		want   float64
		wantOk bool
	}{
		{int(42), 42.0, true},
		{int8(42), 42.0, true},
		{int16(42), 42.0, true},
		{int32(42), 42.0, true},
		{int64(42), 42.0, true},
		{uint(42), 42.0, true},
		{uint8(42), 42.0, true},
		{uint16(42), 42.0, true},
		{uint32(42), 42.0, true},
		{uint64(42), 42.0, true},
		{float32(3.14), 3.140000104904175, true},
		{float64(3.14), 3.14, true},
		{"42", 0, false},
		{nil, 0, false},
		{true, 0, false},
	}

	for _, tt := range tests {
		got, ok := toFloat64(tt.input)
		if ok != tt.wantOk {
			t.Errorf("toFloat64(%v) ok = %v, want %v", tt.input, ok, tt.wantOk)
			continue
		}
		if ok && got != tt.want {
			t.Errorf("toFloat64(%v) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

// === IsEnvVarNameValid tests ===

func TestIsEnvVarNameValidStrict(t *testing.T) {
	valid := []string{"FOO", "FOO_BAR", "FOO123", "_PRIVATE", "A"}
	invalid := []string{"", "123FOO", "FOO-BAR", "FOO.BAR", "FOO BAR", "FOO$BAR"}

	for _, v := range valid {
		if !IsEnvVarNameValid(v) {
			t.Errorf("IsEnvVarNameValid(%q) = false, want true", v)
		}
	}
	for _, v := range invalid {
		if IsEnvVarNameValid(v) {
			t.Errorf("IsEnvVarNameValid(%q) = true, want false", v)
		}
	}
}

// === NormalizeName tests ===

func TestNormalizeNameStrict(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"FOO", "FOO"},
		{"  FOO  ", "FOO"},
		{" FOO BAR ", "FOO BAR"},
	}

	for _, tt := range tests {
		got := NormalizeName(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// === Schema.Validate edge cases ===

func TestSchemaValidateMissingVersion(t *testing.T) {
	s := &Schema{
		Env: map[string]*Variable{
			"FOO": {Type: TypeString},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for missing version")
	}
}

func TestSchemaValidateEmptyEnv(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env:     map[string]*Variable{},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for empty env")
	}
}

func TestSchemaValidateNilVariable(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": nil,
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for nil variable")
	}
}

func TestSchemaValidateInvalidType(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: "invalid"},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for invalid type")
	}
}

func TestSchemaValidateRequiredAndDefault(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString, Required: true, Default: "bar"},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for required and default together")
	}
}

func TestSchemaValidatePatternOnNonString(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeInteger, Pattern: "^[0-9]+$"},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for pattern on non-string")
	}
}

func TestSchemaValidateInvalidPattern(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString, Pattern: "[invalid("},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for invalid regex pattern")
	}
}

func TestSchemaValidateEmptyEnum(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString, Enum: []any{}},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for empty enum")
	}
}

func TestSchemaValidateEnumOnBoolean(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeBoolean, Enum: []any{true, false}},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for enum on boolean")
	}
}

func TestSchemaValidateMinOnString(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString, Min: 10},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for min on string")
	}
}

func TestSchemaValidateMinGreaterThanMax(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeInteger, Min: 100, Max: 10},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for min > max")
	}
}

func TestSchemaValidateMinLengthOnInteger(t *testing.T) {
	minLen := 5
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeInteger, MinLength: &minLen},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for minLength on integer")
	}
}

func TestSchemaValidateMinLengthGreaterThanMaxLength(t *testing.T) {
	minLen := 10
	maxLen := 5
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString, MinLength: &minLen, MaxLength: &maxLen},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for minLength > maxLength")
	}
}

func TestSchemaValidateFormatOnNonString(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeInteger, Format: "email"},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for format on non-string")
	}
}

func TestSchemaValidateUnsupportedFormat(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString, Format: "unsupported"},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestSchemaValidateDisallowOnNonString(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeInteger, Disallow: []string{"1"}},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for disallow on non-string")
	}
}

func TestSchemaValidateDevOnlyAndRequired(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString, DevOnly: true, Required: true},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for devOnly and required together")
	}
}

func TestSchemaValidateDevOnlyAndRequiredIn(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString, DevOnly: true, RequiredIn: []string{"production"}},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for devOnly and requiredIn together")
	}
}

func TestSchemaValidateSeparatorOnNonArray(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString, Separator: ","},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for separator on non-array")
	}
}

func TestSchemaValidateArrayWithoutSeparator(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeArray},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for array without separator")
	}
}

func TestSchemaValidateContainsOnNonArray(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString, Contains: "bar"},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for contains on non-array")
	}
}

func TestSchemaValidateDependsOnWithoutWhen(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString, DependsOn: "BAR"},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for dependsOn without when")
	}
}

func TestSchemaValidateWhenWithoutDependsOn(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString, When: "true"},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for when without dependsOn")
	}
}

func TestSchemaValidateAllowEmptyRedundant(t *testing.T) {
	allowEmpty := false
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString, Required: true, AllowEmpty: &allowEmpty},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for allowEmpty=false with required=true")
	}
}

func TestSchemaValidateTransformOnNonString(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeInteger, Transform: "lowercase"},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for transform on non-string")
	}
}

func TestSchemaValidateUnsupportedTransform(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString, Transform: "reverse"},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for unsupported transform")
	}
}

func TestSchemaValidateInvalidDefaultType(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeInteger, Default: "not-a-number"},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for invalid default type")
	}
}

func TestSchemaValidateCustomSecrets(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString},
		},
		Secrets: &Secrets{
			Custom: []CustomSecretRule{
				{Name: "test", Pattern: "[invalid(", Message: "msg"},
			},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for invalid secret pattern")
	}
}

func TestSchemaValidateCustomSecretMissingName(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString},
		},
		Secrets: &Secrets{
			Custom: []CustomSecretRule{
				{Pattern: "test", Message: "msg"},
			},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for secret rule missing name")
	}
}

func TestSchemaValidateCustomSecretMissingPattern(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString},
		},
		Secrets: &Secrets{
			Custom: []CustomSecretRule{
				{Name: "test", Message: "msg"},
			},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for secret rule missing pattern")
	}
}

// === mergeSchemas tests ===

func TestMergeSchemas(t *testing.T) {
	base := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"BASE": {Type: TypeString},
		},
	}
	child := &Schema{
		Version: "1.1",
		Env: map[string]*Variable{
			"CHILD": {Type: TypeInteger},
		},
	}

	merged := mergeSchemas(base, child)
	if merged.Version != "1.1" {
		t.Errorf("version = %q, want 1.1", merged.Version)
	}
	if _, ok := merged.Env["BASE"]; !ok {
		t.Error("BASE should be inherited")
	}
	if _, ok := merged.Env["CHILD"]; !ok {
		t.Error("CHILD should be present")
	}
}

func TestMergeSchemasChildOverrides(t *testing.T) {
	base := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString, Required: false},
		},
	}
	child := &Schema{
		Version: "1.1",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString, Required: true},
		},
	}

	merged := mergeSchemas(base, child)
	if !merged.Env["FOO"].Required {
		t.Error("child should override base")
	}
}
