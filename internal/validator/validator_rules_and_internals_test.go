package validator

import (
	"strings"
	"testing"

	"github.com/envguard/envguard/internal/schema"
)

// === schemaToInt64 edge cases ===

func TestSchemaToInt64AllTypes(t *testing.T) {
	tests := []struct {
		input  any
		want   int64
		wantOk bool
	}{
		{int(42), 42, true},
		{int8(42), 42, true},
		{int16(42), 42, true},
		{int32(42), 42, true},
		{int64(42), 42, true},
		{uint(42), 42, true},
		{uint8(42), 42, true},
		{uint16(42), 42, true},
		{uint32(42), 42, true},
		{uint64(42), 42, true},
		{float32(42.0), 42, true},
		{float64(42.0), 42, true},
		{float64(42.7), 42, true}, // truncates
		{"42", 0, false},
		{nil, 0, false},
		{true, 0, false},
	}

	for _, tt := range tests {
		got, ok := schemaToInt64(tt.input)
		if ok != tt.wantOk {
			t.Errorf("schemaToInt64(%v) ok = %v, want %v", tt.input, ok, tt.wantOk)
			continue
		}
		if ok && got != tt.want {
			t.Errorf("schemaToInt64(%v) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

// === schemaToFloat64 edge cases ===

func TestSchemaToFloat64AllTypes(t *testing.T) {
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
		{"3.14", 0, false},
		{nil, 0, false},
		{true, 0, false},
	}

	for _, tt := range tests {
		got, ok := schemaToFloat64(tt.input)
		if ok != tt.wantOk {
			t.Errorf("schemaToFloat64(%v) ok = %v, want %v", tt.input, ok, tt.wantOk)
			continue
		}
		if ok && got != tt.want {
			t.Errorf("schemaToFloat64(%v) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

// === int64InSlice edge cases ===

func TestInt64InSlice(t *testing.T) {
	tests := []struct {
		n     int64
		slice []any
		want  bool
	}{
		{2, []any{1, 2, 3}, true},
		{4, []any{1, 2, 3}, false},
		{2, []any{int(2), int64(3), float64(4)}, true},
		{3, []any{int(2), int64(3), float64(4)}, true},
		{4, []any{int(2), int64(3), float64(4)}, true},
		{5, []any{int(2), int64(3), float64(4)}, false},
		{2, []any{"2"}, false},
		{2, []any{}, false},
	}

	for _, tt := range tests {
		got := int64InSlice(tt.n, tt.slice)
		if got != tt.want {
			t.Errorf("int64InSlice(%d, %v) = %v, want %v", tt.n, tt.slice, got, tt.want)
		}
	}
}

// === float64InSlice edge cases ===

func TestFloat64InSlice(t *testing.T) {
	tests := []struct {
		f     float64
		slice []any
		want  bool
	}{
		{2.5, []any{1.5, 2.5, 3.5}, true},
		{4.5, []any{1.5, 2.5, 3.5}, false},
		{2.0, []any{int(2), int64(3), float64(4)}, true},
		{3.0, []any{int(2), int64(3), float64(4)}, true},
		{4.0, []any{int(2), int64(3), float64(4)}, true},
		{5.0, []any{int(2), int64(3), float64(4)}, false},
		{2.5, []any{"2.5"}, false},
		{2.5, []any{}, false},
	}

	for _, tt := range tests {
		got := float64InSlice(tt.f, tt.slice)
		if got != tt.want {
			t.Errorf("float64InSlice(%g, %v) = %v, want %v", tt.f, tt.slice, got, tt.want)
		}
	}
}

// === validateFormat edge cases ===

func TestValidateFormatEmail(t *testing.T) {
	valid := []string{"user@example.com", "a@b.co", "test+tag@sub.domain.com"}
	invalid := []string{"not-an-email", "@example.com", "user@", "Name <user@example.com>"}

	for _, v := range valid {
		if err := validateFormat(v, "email"); err != nil {
			t.Errorf("validateFormat(%q, email) unexpected error: %v", v, err)
		}
	}
	for _, v := range invalid {
		if err := validateFormat(v, "email"); err == nil {
			t.Errorf("validateFormat(%q, email) expected error", v)
		}
	}
}

func TestValidateFormatURL(t *testing.T) {
	valid := []string{"https://example.com", "http://localhost:3000", "ftp://files.example.com"}
	invalid := []string{"not-a-url", "/path/only", "://no-scheme", "http://"}

	for _, v := range valid {
		if err := validateFormat(v, "url"); err != nil {
			t.Errorf("validateFormat(%q, url) unexpected error: %v", v, err)
		}
	}
	for _, v := range invalid {
		if err := validateFormat(v, "url"); err == nil {
			t.Errorf("validateFormat(%q, url) expected error", v)
		}
	}
}

func TestValidateFormatUUID(t *testing.T) {
	valid := []string{"550e8400-e29b-41d4-a716-446655440000", "550E8400-E29B-41D4-A716-446655440000"}
	invalid := []string{"not-a-uuid", "550e8400e29b41d4a716446655440000", "550e8400-e29b-41d4-a716"}

	for _, v := range valid {
		if err := validateFormat(v, "uuid"); err != nil {
			t.Errorf("validateFormat(%q, uuid) unexpected error: %v", v, err)
		}
	}
	for _, v := range invalid {
		if err := validateFormat(v, "uuid"); err == nil {
			t.Errorf("validateFormat(%q, uuid) expected error", v)
		}
	}
}

func TestValidateFormatBase64(t *testing.T) {
	valid := []string{"aGVsbG8=", "d29ybGQ=", ""}
	invalid := []string{"not-valid-base64!!!", "aGVsbG8"}

	for _, v := range valid {
		if err := validateFormat(v, "base64"); err != nil {
			t.Errorf("validateFormat(%q, base64) unexpected error: %v", v, err)
		}
	}
	for _, v := range invalid {
		if err := validateFormat(v, "base64"); err == nil {
			t.Errorf("validateFormat(%q, base64) expected error", v)
		}
	}
}

func TestValidateFormatIP(t *testing.T) {
	valid := []string{"192.168.1.1", "10.0.0.1", "::1", "2001:0db8:85a3:0000:0000:8a2e:0370:7334"}
	invalid := []string{"not-an-ip", "256.1.1.1", ""}

	for _, v := range valid {
		if err := validateFormat(v, "ip"); err != nil {
			t.Errorf("validateFormat(%q, ip) unexpected error: %v", v, err)
		}
	}
	for _, v := range invalid {
		if err := validateFormat(v, "ip"); err == nil {
			t.Errorf("validateFormat(%q, ip) expected error", v)
		}
	}
}

func TestValidateFormatPort(t *testing.T) {
	valid := []string{"1", "80", "8080", "65535"}
	invalid := []string{"0", "65536", "-1", "abc", ""}

	for _, v := range valid {
		if err := validateFormat(v, "port"); err != nil {
			t.Errorf("validateFormat(%q, port) unexpected error: %v", v, err)
		}
	}
	for _, v := range invalid {
		if err := validateFormat(v, "port"); err == nil {
			t.Errorf("validateFormat(%q, port) expected error", v)
		}
	}
}

func TestValidateFormatJSON(t *testing.T) {
	valid := []string{`{"key":"value"}`, `[1,2,3]`, `"string"`, `123`, `true`, `null`}
	invalid := []string{`{invalid}`, ``, `{"key": undefined}`}

	for _, v := range valid {
		if err := validateFormat(v, "json"); err != nil {
			t.Errorf("validateFormat(%q, json) unexpected error: %v", v, err)
		}
	}
	for _, v := range invalid {
		if err := validateFormat(v, "json"); err == nil {
			t.Errorf("validateFormat(%q, json) expected error", v)
		}
	}
}

func TestValidateFormatDuration(t *testing.T) {
	valid := []string{"1h", "30m", "1h30m", "100ms", "-5s"}
	invalid := []string{"not-a-duration", "1", ""}

	for _, v := range valid {
		if err := validateFormat(v, "duration"); err != nil {
			t.Errorf("validateFormat(%q, duration) unexpected error: %v", v, err)
		}
	}
	for _, v := range invalid {
		if err := validateFormat(v, "duration"); err == nil {
			t.Errorf("validateFormat(%q, duration) expected error", v)
		}
	}
}

func TestValidateFormatSemver(t *testing.T) {
	valid := []string{"1.0.0", "0.1.0", "1.2.3-alpha", "1.0.0+build", "1.0.0-alpha.1+build.123"}
	invalid := []string{"1.0", "v1.0.0", "1.0.0.0", "not-semver", ""}

	for _, v := range valid {
		if err := validateFormat(v, "semver"); err != nil {
			t.Errorf("validateFormat(%q, semver) unexpected error: %v", v, err)
		}
	}
	for _, v := range invalid {
		if err := validateFormat(v, "semver"); err == nil {
			t.Errorf("validateFormat(%q, semver) expected error", v)
		}
	}
}

func TestValidateFormatHostname(t *testing.T) {
	valid := []string{"example.com", "localhost", "sub.domain.example.com", "a-b.c"}
	invalid := []string{"-example.com", "example.com-", "", "a..b", "a_b"}

	for _, v := range valid {
		if err := validateFormat(v, "hostname"); err != nil {
			t.Errorf("validateFormat(%q, hostname) unexpected error: %v", v, err)
		}
	}
	for _, v := range invalid {
		if err := validateFormat(v, "hostname"); err == nil {
			t.Errorf("validateFormat(%q, hostname) expected error", v)
		}
	}
}

func TestValidateFormatHex(t *testing.T) {
	valid := []string{"0x1a2b", "1a2b", "0xABCDEF", "1234", "0x0"}
	invalid := []string{"0xGHI", "not-hex", "", "0x"}

	for _, v := range valid {
		if err := validateFormat(v, "hex"); err != nil {
			t.Errorf("validateFormat(%q, hex) unexpected error: %v", v, err)
		}
	}
	for _, v := range invalid {
		if err := validateFormat(v, "hex"); err == nil {
			t.Errorf("validateFormat(%q, hex) expected error", v)
		}
	}
}

func TestValidateFormatCron(t *testing.T) {
	valid := []string{"0 0 * * *", "@daily", "*/5 * * * *", "0 0 1 * *", "@reboot"}
	invalid := []string{"not-cron", "* * *", "", "@invalid"}

	for _, v := range valid {
		if err := validateFormat(v, "cron"); err != nil {
			t.Errorf("validateFormat(%q, cron) unexpected error: %v", v, err)
		}
	}
	for _, v := range invalid {
		if err := validateFormat(v, "cron"); err == nil {
			t.Errorf("validateFormat(%q, cron) expected error", v)
		}
	}
}

func TestValidateFormatUnknown(t *testing.T) {
	// Unknown format should not error (pass-through)
	if err := validateFormat("anything", "unknown_format"); err != nil {
		t.Errorf("validateFormat with unknown format should not error, got: %v", err)
	}
}

// === RedactSensitive edge cases ===

func TestRedactSensitive(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"PASSWORD": {Type: schema.TypeString, Sensitive: true},
			"PUBLIC":   {Type: schema.TypeString, Sensitive: false},
		},
	}

	result := NewResult()
	result.AddError("PASSWORD", "pattern", "value secret123 does not match pattern")
	result.AddError("PUBLIC", "pattern", "value public456 does not match pattern")

	envVars := map[string]string{
		"PASSWORD": "secret123",
		"PUBLIC":   "public456",
	}

	result.RedactSensitive(envVars, s)

	if result.Errors[0].Message != "value *** does not match pattern" {
		t.Errorf("expected sensitive value redacted, got: %s", result.Errors[0].Message)
	}
	if result.Errors[1].Message != "value public456 does not match pattern" {
		t.Errorf("expected non-sensitive value NOT redacted, got: %s", result.Errors[1].Message)
	}
}

func TestRedactSensitiveMissingValue(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"SECRET": {Type: schema.TypeString, Sensitive: true},
		},
	}

	result := NewResult()
	result.AddError("OTHER", "required", "missing")

	envVars := map[string]string{}

	result.RedactSensitive(envVars, s)
	// Should not panic and message should be unchanged
	if result.Errors[0].Message != "missing" {
		t.Errorf("expected unchanged message, got: %s", result.Errors[0].Message)
	}
}

// === defaultToString edge cases ===

func TestDefaultToStringAllTypes(t *testing.T) {
	tests := []struct {
		input any
		want  string
	}{
		{"hello", "hello"},
		{42, "42"},
		{int64(42), "42"},
		{float64(3.14), "3.14"},
		{true, "true"},
		{false, "false"},
		{[]byte("bytes"), "[98 121 116 101 115]"},
		{nil, "<nil>"},
	}

	for _, tt := range tests {
		got := defaultToString(tt.input)
		if got != tt.want {
			t.Errorf("defaultToString(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// === applyTransform edge cases ===

func TestApplyTransform(t *testing.T) {
	tests := []struct {
		value     string
		transform string
		want      string
	}{
		{"Hello", "lowercase", "hello"},
		{"Hello", "uppercase", "HELLO"},
		{"  Hello  ", "trim", "Hello"},
		{"Hello", "", "Hello"},
		{"Hello", "unknown", "Hello"},
	}

	for _, tt := range tests {
		got := applyTransform(tt.value, tt.transform)
		if got != tt.want {
			t.Errorf("applyTransform(%q, %q) = %q, want %q", tt.value, tt.transform, got, tt.want)
		}
	}
}

// === stringInSlice edge cases ===

func TestStringInSlice(t *testing.T) {
	tests := []struct {
		s     string
		slice []any
		want  bool
	}{
		{"a", []any{"a", "b", "c"}, true},
		{"d", []any{"a", "b", "c"}, false},
		{"a", []any{}, false},
		{"a", []any{1, 2, 3}, false},
	}

	for _, tt := range tests {
		got := stringInSlice(tt.s, tt.slice)
		if got != tt.want {
			t.Errorf("stringInSlice(%q, %v) = %v, want %v", tt.s, tt.slice, got, tt.want)
		}
	}
}

// === validateVariable edge cases ===

func TestValidateVariableDeprecated(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"OLD_VAR": {Type: schema.TypeString, Deprecated: "Use NEW_VAR instead"},
		},
	}
	vars := map[string]string{"OLD_VAR": "value"}
	result := Validate(s, vars, false, "")
	if len(result.Warnings) != 1 {
		t.Errorf("expected 1 deprecation warning, got %d", len(result.Warnings))
	}
}

func TestValidateVariableDevOnly(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"DEBUG": {Type: schema.TypeString, DevOnly: true},
		},
	}

	// In dev environment, should be required
	vars := map[string]string{}
	result := Validate(s, vars, false, "development")
	if result.Valid {
		t.Error("expected validation to fail for missing devOnly var in development")
	}

	// In production, should be ignored
	result = Validate(s, vars, false, "production")
	if !result.Valid {
		t.Errorf("expected validation to pass in production, got errors: %v", result.Errors)
	}

	// With no env-name, dev-only is treated as dev
	result = Validate(s, vars, false, "")
	if result.Valid {
		t.Error("expected validation to fail for missing devOnly var with empty envName")
	}
}

func TestValidateVariableRequiredIn(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"DB_PASSWORD": {Type: schema.TypeString, RequiredIn: []string{"production", "staging"}},
		},
	}

	// Required in production
	vars := map[string]string{}
	result := Validate(s, vars, false, "production")
	if result.Valid {
		t.Error("expected validation to fail in production")
	}

	// Not required in development
	result = Validate(s, vars, false, "development")
	if !result.Valid {
		t.Errorf("expected validation to pass in development, got errors: %v", result.Errors)
	}

	// Required in staging (case-insensitive)
	result = Validate(s, vars, false, "Staging")
	if result.Valid {
		t.Error("expected validation to fail in Staging")
	}
}

func TestValidateVariableDependsOn(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"HTTPS":    {Type: schema.TypeBoolean},
			"SSL_CERT": {Type: schema.TypeString, DependsOn: "HTTPS", When: "true"},
		},
	}

	// When HTTPS=true, SSL_CERT is required
	vars := map[string]string{"HTTPS": "true"}
	result := Validate(s, vars, false, "")
	if result.Valid {
		t.Error("expected validation to fail when dependsOn triggers")
	}

	// When HTTPS=false, SSL_CERT is not required
	vars = map[string]string{"HTTPS": "false"}
	result = Validate(s, vars, false, "")
	if !result.Valid {
		t.Errorf("expected validation to pass when dependsOn does not trigger, got errors: %v", result.Errors)
	}

	// When HTTPS missing, SSL_CERT is not required
	vars = map[string]string{}
	result = Validate(s, vars, false, "")
	if !result.Valid {
		t.Errorf("expected validation to pass when dependsOn missing, got errors: %v", result.Errors)
	}
}

func TestValidateVariableAllowEmptyFalse(t *testing.T) {
	allowEmpty := false
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"OPTIONAL": {Type: schema.TypeString, AllowEmpty: &allowEmpty},
		},
	}

	vars := map[string]string{"OPTIONAL": ""}
	result := Validate(s, vars, false, "")
	if result.Valid {
		t.Error("expected validation to fail for empty value when allowEmpty=false")
	}

	vars = map[string]string{"OPTIONAL": "value"}
	result = Validate(s, vars, false, "")
	if !result.Valid {
		t.Errorf("expected validation to pass for non-empty value, got errors: %v", result.Errors)
	}
}

func TestValidateVariableTransform(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"LOWER": {Type: schema.TypeString, Transform: "lowercase"},
			"UPPER": {Type: schema.TypeString, Transform: "uppercase"},
			"TRIM":  {Type: schema.TypeString, Transform: "trim"},
			"ENUM":  {Type: schema.TypeString, Transform: "lowercase", Enum: []any{"a", "b", "c"}},
		},
	}

	vars := map[string]string{
		"LOWER": "Hello",
		"UPPER": "Hello",
		"TRIM":  "  hello  ",
		"ENUM":  "A",
	}
	result := Validate(s, vars, false, "")
	if !result.Valid {
		t.Errorf("expected validation to pass after transform, got errors: %v", result.Errors)
	}
}

func TestValidateVariableDisallow(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"ENV": {Type: schema.TypeString, Disallow: []string{"production", "staging"}},
		},
	}

	vars := map[string]string{"ENV": "production"}
	result := Validate(s, vars, false, "")
	if result.Valid {
		t.Error("expected validation to fail for disallowed value")
	}

	vars = map[string]string{"ENV": "development"}
	result = Validate(s, vars, false, "")
	if !result.Valid {
		t.Errorf("expected validation to pass for allowed value, got errors: %v", result.Errors)
	}
}

func TestValidateVariableMinLengthMaxLength(t *testing.T) {
	minLen := 3
	maxLen := 5
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"CODE": {Type: schema.TypeString, MinLength: &minLen, MaxLength: &maxLen},
		},
	}

	vars := map[string]string{"CODE": "ab"}
	result := Validate(s, vars, false, "")
	if result.Valid {
		t.Error("expected validation to fail for value below minLength")
	}

	vars = map[string]string{"CODE": "abcdef"}
	result = Validate(s, vars, false, "")
	if result.Valid {
		t.Error("expected validation to fail for value above maxLength")
	}

	vars = map[string]string{"CODE": "abcd"}
	result = Validate(s, vars, false, "")
	if !result.Valid {
		t.Errorf("expected validation to pass for value within range, got errors: %v", result.Errors)
	}
}

func TestValidateVariableIntegerMinMax(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"PORT": {Type: schema.TypeInteger, Min: 1024, Max: 65535},
		},
	}

	vars := map[string]string{"PORT": "80"}
	result := Validate(s, vars, false, "")
	if result.Valid {
		t.Error("expected validation to fail for value below min")
	}

	vars = map[string]string{"PORT": "70000"}
	result = Validate(s, vars, false, "")
	if result.Valid {
		t.Error("expected validation to fail for value above max")
	}

	vars = map[string]string{"PORT": "8080"}
	result = Validate(s, vars, false, "")
	if !result.Valid {
		t.Errorf("expected validation to pass for value within range, got errors: %v", result.Errors)
	}
}

func TestValidateVariableFloatMinMax(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"RATIO": {Type: schema.TypeFloat, Min: 0.0, Max: 1.0},
		},
	}

	vars := map[string]string{"RATIO": "-0.5"}
	result := Validate(s, vars, false, "")
	if result.Valid {
		t.Error("expected validation to fail for value below min")
	}

	vars = map[string]string{"RATIO": "1.5"}
	result = Validate(s, vars, false, "")
	if result.Valid {
		t.Error("expected validation to fail for value above max")
	}

	vars = map[string]string{"RATIO": "0.5"}
	result = Validate(s, vars, false, "")
	if !result.Valid {
		t.Errorf("expected validation to pass for value within range, got errors: %v", result.Errors)
	}
}

func TestValidateVariableArrayContains(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"ROLES": {Type: schema.TypeArray, Separator: ",", Contains: "admin"},
		},
	}

	vars := map[string]string{"ROLES": "read,write"}
	result := Validate(s, vars, false, "")
	if result.Valid {
		t.Error("expected validation to fail when array does not contain required item")
	}

	vars = map[string]string{"ROLES": "read,admin,write"}
	result = Validate(s, vars, false, "")
	if !result.Valid {
		t.Errorf("expected validation to pass when array contains item, got errors: %v", result.Errors)
	}
}

func TestValidateVariableArrayMinMaxLength(t *testing.T) {
	minLen := 2
	maxLen := 4
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"ITEMS": {Type: schema.TypeArray, Separator: ",", MinLength: &minLen, MaxLength: &maxLen},
		},
	}

	vars := map[string]string{"ITEMS": "a"}
	result := Validate(s, vars, false, "")
	if result.Valid {
		t.Error("expected validation to fail for array below minLength")
	}

	vars = map[string]string{"ITEMS": "a,b,c,d,e"}
	result = Validate(s, vars, false, "")
	if result.Valid {
		t.Error("expected validation to fail for array above maxLength")
	}

	vars = map[string]string{"ITEMS": "a,b,c"}
	result = Validate(s, vars, false, "")
	if !result.Valid {
		t.Errorf("expected validation to pass for array within range, got errors: %v", result.Errors)
	}
}

func TestValidateVariableEmptyArray(t *testing.T) {
	// Directly test validateArray with empty string since Validate() returns early for optional empty values
	result := NewResult()
	v := &schema.Variable{Type: schema.TypeArray, Separator: ","}
	validateArray(result, "ITEMS", v, "")
	if result.Valid {
		t.Error("expected validation to fail for empty array string")
	}
	if !strings.Contains(result.Errors[0].Message, "empty string") {
		t.Errorf("expected empty string error, got: %s", result.Errors[0].Message)
	}
}
