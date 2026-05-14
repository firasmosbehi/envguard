package validator

import (
	"testing"

	"github.com/envguard/envguard/internal/schema"
)

// === ValidateParallel tests ===

func TestValidateParallel(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"VAR1": {Type: schema.TypeString, Required: true},
			"VAR2": {Type: schema.TypeInteger, Required: true},
			"VAR3": {Type: schema.TypeBoolean, Required: true},
		},
	}

	envVars := map[string]string{
		"VAR1": "hello",
		"VAR2": "42",
		"VAR3": "true",
	}

	result := ValidateParallel(s, envVars, false, "", true)
	if !result.Valid {
		t.Errorf("expected valid result, got errors: %v", result.Errors)
	}
}

func TestValidateParallelWithErrors(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"VAR1": {Type: schema.TypeString, Required: true},
			"VAR2": {Type: schema.TypeInteger, Required: true},
		},
	}

	envVars := map[string]string{
		"VAR1": "",
		"VAR2": "not-a-number",
	}

	result := ValidateParallel(s, envVars, false, "", true)
	if result.Valid {
		t.Error("expected invalid result")
	}
	if len(result.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(result.Errors))
	}
}

func TestValidateParallelSingleVar(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"VAR1": {Type: schema.TypeString, Required: true},
		},
	}

	envVars := map[string]string{"VAR1": "hello"}
	result := ValidateParallel(s, envVars, false, "", true)
	if !result.Valid {
		t.Error("expected valid result")
	}
}

// === mergeResults tests ===

func TestMergeResults(t *testing.T) {
	dst := NewResult()
	src := NewResult()
	src.Valid = false
	src.AddError("KEY1", "required", "missing")
	src.AddWarning("KEY2", "strict", "unknown")

	mergeResults(dst, src)

	if dst.Valid {
		t.Error("expected dst to be invalid after merge")
	}
	if len(dst.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(dst.Errors))
	}
	if len(dst.Warnings) != 1 {
		t.Errorf("expected 1 warning, got %d", len(dst.Warnings))
	}
}

// === HasErrors tests ===

func TestHasErrors(t *testing.T) {
	r := NewResult()
	r.AddErrorWithSeverity("KEY", "rule", "msg", SeverityWarn)

	if !r.HasErrors(SeverityWarn) {
		t.Error("expected HasErrors(SeverityWarn) to be true")
	}
	if r.HasErrors(SeverityError) {
		t.Error("expected HasErrors(SeverityError) to be false because warn < error")
	}
	if r.HasErrors(SeverityInfo) {
		t.Error("expected HasErrors(SeverityInfo) to be false")
	}
}

func TestHasErrorsNoErrors(t *testing.T) {
	r := NewResult()
	if r.HasErrors(SeverityError) {
		t.Error("expected HasErrors to be false for empty result")
	}
}

// === RedactSensitive tests ===

func TestRedactSensitiveValueNotExists(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"SECRET": {Type: schema.TypeString, Sensitive: true},
		},
	}
	r := NewResult()
	r.AddWarning("OTHER", "rule", "secret value here")
	r.RedactSensitive(map[string]string{}, s)
	if !r.Valid {
		t.Error("result should still be valid")
	}
}

func TestRedactSensitiveValueEmpty(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"SECRET": {Type: schema.TypeString, Sensitive: true},
		},
	}
	r := NewResult()
	r.AddError("OTHER", "rule", "something")
	r.RedactSensitive(map[string]string{"SECRET": ""}, s)
	if len(r.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(r.Errors))
	}
}

// === RegexCache tests ===

func TestRegexCacheClear(t *testing.T) {
	c := &RegexCache{}
	_, err := c.Compile("^[a-z]+$")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	c.Clear()
	// After clear, it should still compile (just not from cache)
	_, err = c.Compile("^[a-z]+$")
	if err != nil {
		t.Fatalf("Compile after clear failed: %v", err)
	}
}

func TestRegexCacheCompileError(t *testing.T) {
	c := &RegexCache{}
	_, err := c.Compile("[invalid(")
	if err == nil {
		t.Error("expected error for invalid regex")
	}
}

// === validateString edge cases ===

func TestValidateStringPrefixFail(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"FOO": {Type: schema.TypeString, Prefix: "pre"},
		},
	}
	r := Validate(s, map[string]string{"FOO": "fix"}, false, "")
	if r.Valid {
		t.Error("expected error for prefix mismatch")
	}
}

func TestValidateStringSuffixFail(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"FOO": {Type: schema.TypeString, Suffix: "suf"},
		},
	}
	r := Validate(s, map[string]string{"FOO": "fus"}, false, "")
	if r.Valid {
		t.Error("expected error for suffix mismatch")
	}
}

func TestValidateStringFormatFail(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"EMAIL": {Type: schema.TypeString, Format: "email"},
		},
	}
	r := Validate(s, map[string]string{"EMAIL": "not-an-email"}, false, "")
	if r.Valid {
		t.Error("expected error for invalid email format")
	}
}

func TestValidateStringDisallowFail(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"FOO": {Type: schema.TypeString, Disallow: []string{"bad", "worse"}},
		},
	}
	r := Validate(s, map[string]string{"FOO": "bad"}, false, "")
	if r.Valid {
		t.Error("expected error for disallowed value")
	}
}

func TestValidateStringPatternFail(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"FOO": {Type: schema.TypeString, Pattern: "^[0-9]+$"},
		},
	}
	r := Validate(s, map[string]string{"FOO": "abc"}, false, "")
	if r.Valid {
		t.Error("expected error for pattern mismatch")
	}
}

func TestValidateStringMinLengthFail(t *testing.T) {
	minLen := 5
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"FOO": {Type: schema.TypeString, MinLength: &minLen},
		},
	}
	r := Validate(s, map[string]string{"FOO": "ab"}, false, "")
	if r.Valid {
		t.Error("expected error for minLength violation")
	}
}

func TestValidateStringMaxLengthFail(t *testing.T) {
	maxLen := 3
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"FOO": {Type: schema.TypeString, MaxLength: &maxLen},
		},
	}
	r := Validate(s, map[string]string{"FOO": "abcd"}, false, "")
	if r.Valid {
		t.Error("expected error for maxLength violation")
	}
}

// === validateInteger edge cases ===

func TestValidateIntegerMinFail(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"COUNT": {Type: schema.TypeInteger, Min: 10},
		},
	}
	r := Validate(s, map[string]string{"COUNT": "5"}, false, "")
	if r.Valid {
		t.Error("expected error for min violation")
	}
}

func TestValidateIntegerMaxFail(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"COUNT": {Type: schema.TypeInteger, Max: 100},
		},
	}
	r := Validate(s, map[string]string{"COUNT": "200"}, false, "")
	if r.Valid {
		t.Error("expected error for max violation")
	}
}

func TestValidateIntegerMultipleOfFail(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"COUNT": {Type: schema.TypeInteger, MultipleOf: 5},
		},
	}
	r := Validate(s, map[string]string{"COUNT": "7"}, false, "")
	if r.Valid {
		t.Error("expected error for multipleOf violation")
	}
}

// === validateFloat edge cases ===

func TestValidateFloatMinFail(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"RATIO": {Type: schema.TypeFloat, Min: 0.0},
		},
	}
	r := Validate(s, map[string]string{"RATIO": "-1.5"}, false, "")
	if r.Valid {
		t.Error("expected error for min violation")
	}
}

func TestValidateFloatMaxFail(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"RATIO": {Type: schema.TypeFloat, Max: 1.0},
		},
	}
	r := Validate(s, map[string]string{"RATIO": "2.5"}, false, "")
	if r.Valid {
		t.Error("expected error for max violation")
	}
}

func TestValidateFloatMultipleOfFail(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"RATIO": {Type: schema.TypeFloat, MultipleOf: 0.5},
		},
	}
	r := Validate(s, map[string]string{"RATIO": "0.7"}, false, "")
	if r.Valid {
		t.Error("expected error for multipleOf violation")
	}
}

// === validateArray edge cases ===

func TestValidateArrayNotEmptyFail(t *testing.T) {
	notEmpty := true
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"TAGS": {Type: schema.TypeArray, Separator: ",", NotEmpty: &notEmpty},
		},
	}
	r := Validate(s, map[string]string{"TAGS": ""}, false, "")
	if r.Valid {
		t.Error("expected error for empty array with notEmpty")
	}
}

func TestValidateArrayUniqueItemsFail(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"TAGS": {Type: schema.TypeArray, Separator: ",", UniqueItems: true},
		},
	}
	r := Validate(s, map[string]string{"TAGS": "a,b,a"}, false, "")
	if r.Valid {
		t.Error("expected error for duplicate items")
	}
}

func TestValidateArrayItemPatternFail(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"TAGS": {Type: schema.TypeArray, Separator: ",", ItemPattern: "^[a-z]+$"},
		},
	}
	r := Validate(s, map[string]string{"TAGS": "abc,DEF"}, false, "")
	if r.Valid {
		t.Error("expected error for itemPattern mismatch")
	}
}

func TestValidateArrayItemTypeIntegerFail(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"NUMS": {Type: schema.TypeArray, Separator: ",", ItemType: schema.TypeInteger},
		},
	}
	r := Validate(s, map[string]string{"NUMS": "1,not-a-number"}, false, "")
	if r.Valid {
		t.Error("expected error for invalid itemType")
	}
}

func TestValidateArrayItemTypeFloatFail(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"NUMS": {Type: schema.TypeArray, Separator: ",", ItemType: schema.TypeFloat},
		},
	}
	r := Validate(s, map[string]string{"NUMS": "1.5,not-a-number"}, false, "")
	if r.Valid {
		t.Error("expected error for invalid itemType")
	}
}

func TestValidateArrayItemTypeBooleanFail(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"FLAGS": {Type: schema.TypeArray, Separator: ",", ItemType: schema.TypeBoolean},
		},
	}
	r := Validate(s, map[string]string{"FLAGS": "true,maybe"}, false, "")
	if r.Valid {
		t.Error("expected error for invalid itemType")
	}
}

func TestValidateArrayContainsFail(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"TAGS": {Type: schema.TypeArray, Separator: ",", Contains: "required"},
		},
	}
	r := Validate(s, map[string]string{"TAGS": "a,b,c"}, false, "")
	if r.Valid {
		t.Error("expected error for missing required item")
	}
}

// === validateFormat remaining formats ===

func TestValidateFormatDatetime(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"DT": {Type: schema.TypeString, Format: "datetime"},
		},
	}
	r := Validate(s, map[string]string{"DT": "2024-01-15T10:30:00Z"}, false, "")
	if !r.Valid {
		t.Errorf("expected valid datetime, got: %v", r.Errors)
	}

	r = Validate(s, map[string]string{"DT": "not-a-datetime"}, false, "")
	if r.Valid {
		t.Error("expected error for invalid datetime")
	}
}

func TestValidateFormatDate(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"D": {Type: schema.TypeString, Format: "date"},
		},
	}
	r := Validate(s, map[string]string{"D": "2024-01-15"}, false, "")
	if !r.Valid {
		t.Errorf("expected valid date, got: %v", r.Errors)
	}

	r = Validate(s, map[string]string{"D": "15-01-2024"}, false, "")
	if r.Valid {
		t.Error("expected error for invalid date")
	}
}

func TestValidateFormatTime(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"T": {Type: schema.TypeString, Format: "time"},
		},
	}
	r := Validate(s, map[string]string{"T": "10:30:00"}, false, "")
	if !r.Valid {
		t.Errorf("expected valid time, got: %v", r.Errors)
	}

	r = Validate(s, map[string]string{"T": "25:00:00"}, false, "")
	if r.Valid {
		t.Error("expected error for invalid time")
	}
}

func TestValidateFormatTimezone(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"TZ": {Type: schema.TypeString, Format: "timezone"},
		},
	}
	r := Validate(s, map[string]string{"TZ": "America/New_York"}, false, "")
	if !r.Valid {
		t.Errorf("expected valid timezone, got: %v", r.Errors)
	}

	r = Validate(s, map[string]string{"TZ": "Mars/Colony"}, false, "")
	if r.Valid {
		t.Error("expected error for invalid timezone")
	}
}

func TestValidateFormatColor(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"COLOR": {Type: schema.TypeString, Format: "color"},
		},
	}

	validColors := []string{"#fff", "#ffffff", "rgb(255,0,0)", "rgba(255,0,0,0.5)", "hsl(120,50%,50%)"}
	for _, c := range validColors {
		r := Validate(s, map[string]string{"COLOR": c}, false, "")
		if !r.Valid {
			t.Errorf("expected valid color %q, got: %v", c, r.Errors)
		}
	}

	r := Validate(s, map[string]string{"COLOR": "not-a-color"}, false, "")
	if r.Valid {
		t.Error("expected error for invalid color")
	}
}

func TestValidateFormatSlug(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"SLUG": {Type: schema.TypeString, Format: "slug"},
		},
	}
	r := Validate(s, map[string]string{"SLUG": "my-slug"}, false, "")
	if !r.Valid {
		t.Errorf("expected valid slug, got: %v", r.Errors)
	}

	r = Validate(s, map[string]string{"SLUG": "My Slug"}, false, "")
	if r.Valid {
		t.Error("expected error for invalid slug")
	}
}

func TestValidateFormatFilepath(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"PATH": {Type: schema.TypeString, Format: "filepath"},
		},
	}
	r := Validate(s, map[string]string{"PATH": "/tmp/file.txt"}, false, "")
	if !r.Valid {
		t.Errorf("expected valid filepath, got: %v", r.Errors)
	}

	r = Validate(s, map[string]string{"PATH": "file\x00name"}, false, "")
	if r.Valid {
		t.Error("expected error for filepath with null bytes")
	}
}

func TestValidateFormatDirectory(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"DIR": {Type: schema.TypeString, Format: "directory"},
		},
	}
	r := Validate(s, map[string]string{"DIR": "/tmp"}, false, "")
	if !r.Valid {
		t.Errorf("expected valid directory, got: %v", r.Errors)
	}

	r = Validate(s, map[string]string{"DIR": "dir\x00name"}, false, "")
	if r.Valid {
		t.Error("expected error for directory with null bytes")
	}
}

func TestValidateFormatLocale(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"LOC": {Type: schema.TypeString, Format: "locale"},
		},
	}
	r := Validate(s, map[string]string{"LOC": "en-US"}, false, "")
	if !r.Valid {
		t.Errorf("expected valid locale, got: %v", r.Errors)
	}

	r = Validate(s, map[string]string{"LOC": "invalid-locale-123"}, false, "")
	if r.Valid {
		t.Error("expected error for invalid locale")
	}
}

func TestValidateFormatJWT(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"TOKEN": {Type: schema.TypeString, Format: "jwt"},
		},
	}
	r := Validate(s, map[string]string{"TOKEN": "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.signature"}, false, "")
	if !r.Valid {
		t.Errorf("expected valid jwt, got: %v", r.Errors)
	}

	r = Validate(s, map[string]string{"TOKEN": "not-a-jwt"}, false, "")
	if r.Valid {
		t.Error("expected error for invalid jwt")
	}
}

func TestValidateFormatMongoDBURI(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"MONGO": {Type: schema.TypeString, Format: "mongodb-uri"},
		},
	}
	r := Validate(s, map[string]string{"MONGO": "mongodb://localhost:27017/db"}, false, "")
	if !r.Valid {
		t.Errorf("expected valid mongodb-uri, got: %v", r.Errors)
	}

	r = Validate(s, map[string]string{"MONGO": "not-a-mongodb-uri"}, false, "")
	if r.Valid {
		t.Error("expected error for invalid mongodb-uri")
	}
}

func TestValidateFormatRedisURI(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"REDIS": {Type: schema.TypeString, Format: "redis-uri"},
		},
	}
	r := Validate(s, map[string]string{"REDIS": "redis://localhost:6379"}, false, "")
	if !r.Valid {
		t.Errorf("expected valid redis-uri, got: %v", r.Errors)
	}

	r = Validate(s, map[string]string{"REDIS": "not-a-redis-uri"}, false, "")
	if r.Valid {
		t.Error("expected error for invalid redis-uri")
	}
}

// === Result.IsValid with failOnWarnings ===

func TestResultIsValidFailOnWarnings(t *testing.T) {
	r := NewResult()
	r.AddErrorWithSeverity("KEY", "rule", "msg", SeverityWarn)

	if !r.IsValid(false) {
		t.Error("expected IsValid(false) to be true when only warnings")
	}
	if r.IsValid(true) {
		t.Error("expected IsValid(true) to be false when failOnWarnings")
	}
}
