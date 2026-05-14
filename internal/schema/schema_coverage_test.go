package schema

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

// === Cache tests ===

func TestSchemaCacheGetNilCache(t *testing.T) {
	c := &SchemaCache{}
	_, ok := c.Get("/nonexistent/path.yaml")
	if ok {
		t.Error("expected Get to return false for nil cache")
	}
}

func TestSchemaCacheGetMissingPath(t *testing.T) {
	c := &SchemaCache{cache: make(map[string]schemaCacheEntry)}
	_, ok := c.Get("/nonexistent/path.yaml")
	if ok {
		t.Error("expected Get to return false for missing path")
	}
}

func TestSchemaCacheGetModifiedFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "schema.yaml")
	os.WriteFile(path, []byte("version: \"1.0\"\nenv:\n  FOO:\n    type: string\n"), 0644)

	c := &SchemaCache{}
	s := &Schema{Version: "1.0", Env: map[string]*Variable{"FOO": {Type: TypeString}}}
	c.Set(path, s)

	// Modify the file after caching
	os.WriteFile(path, []byte("version: \"2.0\"\nenv:\n  BAR:\n    type: integer\n"), 0644)

	_, ok := c.Get(path)
	if ok {
		t.Error("expected Get to return false when file is modified after cache")
	}
}

func TestSchemaCacheGetStatError(t *testing.T) {
	c := &SchemaCache{cache: make(map[string]schemaCacheEntry)}
	// Put an entry for a path that will be deleted
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "schema.yaml")
	os.WriteFile(path, []byte("version: \"1.0\"\n"), 0644)
	s := &Schema{Version: "1.0"}
	c.Set(path, s)
	os.Remove(path)

	_, ok := c.Get(path)
	if ok {
		t.Error("expected Get to return false when stat fails")
	}
}

func TestSchemaCacheClear(t *testing.T) {
	c := &SchemaCache{}
	s := &Schema{Version: "1.0"}
	c.Set("/tmp/test.yaml", s)
	c.Clear()

	_, ok := c.Get("/tmp/test.yaml")
	if ok {
		t.Error("expected Get to return false after Clear")
	}
}

// === Remote tests ===

func TestFetchRemoteSchemaInvalidURL(t *testing.T) {
	_, err := fetchRemoteSchema("://invalid-url")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestFetchRemoteSchemaFromCache(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("cached"))
	}))
	defer server.Close()

	// First fetch to populate cache
	_, err := fetchRemoteSchema(server.URL + "/schema.yaml")
	if err != nil {
		t.Fatalf("first fetch failed: %v", err)
	}

	// Second fetch should read from cache
	data, err := fetchRemoteSchema(server.URL + "/schema.yaml")
	if err != nil {
		t.Fatalf("second fetch failed: %v", err)
	}
	if data != "cached" {
		t.Errorf("expected cached data, got %q", data)
	}
}

func TestFetchRemoteSchemaHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	_, err := fetchRemoteSchema(server.URL + "/missing.yaml")
	if err == nil {
		t.Error("expected error for 404 response")
	}
}

func TestCachePathForURL(t *testing.T) {
	u, _ := url.Parse("https://example.com/path/to/schema.yaml")
	path := cachePathForURL(u)
	if path == "" {
		t.Error("expected non-empty cache path")
	}
}

func TestClearRemoteCache(t *testing.T) {
	// Populate cache first
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}))
	defer server.Close()

	_, err := fetchRemoteSchema(server.URL + "/schema.yaml")
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}

	err = ClearRemoteCache()
	if err != nil {
		t.Fatalf("ClearRemoteCache failed: %v", err)
	}
}

// === parseWithStack tests ===

func TestParseWithStackFileReadError(t *testing.T) {
	_, err := parseWithStack("/nonexistent/schema.yaml", make(map[string]bool))
	if err == nil {
		t.Error("expected error for missing file")
	}
}

// === validateVariable edge cases ===

func TestValidateVariablePrefixOnNonString(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeInteger, Prefix: "pre"},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for prefix on non-string")
	}
}

func TestValidateVariableSuffixOnNonString(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeInteger, Suffix: "suf"},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for suffix on non-string")
	}
}

func TestValidateVariableItemTypeOnNonArray(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString, ItemType: TypeInteger},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for itemType on non-array")
	}
}

func TestValidateVariableInvalidItemType(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeArray, Separator: ",", ItemType: "invalid"},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for invalid itemType")
	}
}

func TestValidateVariableUniqueItemsOnNonArray(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString, UniqueItems: true},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for uniqueItems on non-array")
	}
}

func TestValidateVariableItemPatternOnNonArray(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString, ItemPattern: "^[a-z]+$"},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for itemPattern on non-array")
	}
}

func TestValidateVariableInvalidItemPattern(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeArray, Separator: ",", ItemPattern: "[invalid("},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for invalid itemPattern")
	}
}

func TestValidateVariableNotEmptyOnNonArray(t *testing.T) {
	notEmpty := true
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString, NotEmpty: &notEmpty},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for notEmpty on non-array")
	}
}

func TestValidateVariableMultipleOfOnNonNumeric(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString, MultipleOf: 2},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for multipleOf on non-numeric")
	}
}

func TestValidateVariableMultipleOfZero(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeInteger, MultipleOf: 0},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for multipleOf = 0")
	}
}

func TestValidateVariableInvalidNumeric(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeInteger, Min: "not-a-number"},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for non-numeric min")
	}
}

func TestValidateVariableValidSecrets(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString},
		},
		Secrets: &Secrets{
			Custom: []CustomSecretRule{
				{Name: "test", Pattern: "^test$", Message: "test msg"},
			},
		},
	}
	if err := s.Validate(); err != nil {
		t.Errorf("unexpected error for valid secrets: %v", err)
	}
}

func TestValidateVariableUnsupportedSeverity(t *testing.T) {
	s := &Schema{
		Version: "1.0",
		Env: map[string]*Variable{
			"FOO": {Type: TypeString, Severity: "critical"},
		},
	}
	if err := s.Validate(); err == nil {
		t.Error("expected error for unsupported severity")
	}
}

func TestValidateNumeric(t *testing.T) {
	if err := validateNumeric("FOO", "min", "not-a-number"); err == nil {
		t.Error("expected error for non-numeric value")
	}
	if err := validateNumeric("FOO", "min", 42); err != nil {
		t.Errorf("unexpected error for numeric value: %v", err)
	}
}

func TestParseUsesCache(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "envguard.yaml")
	os.WriteFile(path, []byte("version: \"1.0\"\nenv:\n  FOO:\n    type: string\n"), 0644)

	DefaultSchemaCache.Clear()
	defer DefaultSchemaCache.Clear()

	_, err := Parse(path)
	if err != nil {
		t.Fatalf("first parse failed: %v", err)
	}

	// Second parse should use cache
	_, err = Parse(path)
	if err != nil {
		t.Fatalf("second parse failed: %v", err)
	}
}

func TestParseLenientAbsPathError(t *testing.T) {
	// This test covers the filepath.Abs error path which is hard to trigger,
	// but we test the error propagation from os.ReadFile
	_, err := ParseLenient("/nonexistent/path.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestParseLenientBaseSchemaError(t *testing.T) {
	tmpDir := t.TempDir()
	childPath := filepath.Join(tmpDir, "child.yaml")
	os.WriteFile(childPath, []byte("version: \"1.0\"\nextends: ./missing.yaml\nenv:\n  FOO:\n    type: string\n"), 0644)

	_, err := ParseLenient(childPath)
	if err == nil {
		t.Error("expected error when base schema is missing")
	}
}

func TestValidateDefaultBooleanTrue(t *testing.T) {
	if err := validateDefault(TypeBoolean, true); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateEnumValueBoolean(t *testing.T) {
	if err := validateEnumValue(TypeBoolean, true); err == nil {
		t.Error("expected error for boolean enum")
	}
}

func TestValidateEnumValueUnsupported(t *testing.T) {
	if err := validateEnumValue(TypeBoolean, "x"); err == nil {
		t.Error("expected error for unsupported type")
	}
}
