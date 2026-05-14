package docs

import (
	"strings"
	"testing"

	"github.com/envguard/envguard/internal/schema"
)

func TestGenerateMarkdownGroupByPrefix(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"DB_HOST":   {Type: schema.TypeString, Required: true, Description: "Database host"},
			"DB_PORT":   {Type: schema.TypeInteger, Default: 5432},
			"API_KEY":   {Type: schema.TypeString, Sensitive: true},
			"REDIS_URL": {Type: schema.TypeString, Format: "url"},
			"NOPREFIX":  {Type: schema.TypeString},
		},
	}

	out, err := Generate(s, Options{Format: "markdown", GroupBy: "prefix"})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(out, "## DB") {
		t.Error("expected DB group header")
	}
	if !strings.Contains(out, "## API") {
		t.Error("expected API group header")
	}
	if !strings.Contains(out, "## REDIS") {
		t.Error("expected REDIS group header")
	}
	if !strings.Contains(out, "## Other") {
		t.Error("expected Other group header for NOPREFIX")
	}
}

func TestGenerateMarkdownWithAllFields(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"FULL_VAR": {
				Type:        schema.TypeString,
				Required:    true,
				Default:     "default",
				Description: "A full variable",
				Enum:        []any{"a", "b"},
				Format:      "email",
				Pattern:     "^[a-z]+$",
				Min:         1,
				Max:         10,
				Deprecated:  "Use NEW_VAR instead",
				Sensitive:   true,
			},
		},
	}

	out, err := Generate(s, Options{Format: "markdown"})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	checks := []string{"Required", "Default", "Enum", "Format", "Pattern", "Min", "Max", "Deprecated", "Sensitive"}
	for _, c := range checks {
		if !strings.Contains(out, c) {
			t.Errorf("expected output to contain %s", c)
		}
	}
}

func TestGenerateHTMLEmptySchema(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"FOO": {Type: schema.TypeString},
		},
	}

	out, err := Generate(s, Options{Format: "html"})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(out, "<!DOCTYPE html>") {
		t.Error("expected HTML doctype")
	}
	if !strings.Contains(out, "FOO") {
		t.Error("expected FOO in output")
	}
}

func TestGenerateJSONWithMultipleVars(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"REQUIRED_VAR": {
				Type:     schema.TypeString,
				Required: true,
			},
			"OPTIONAL_VAR": {
				Type:        schema.TypeString,
				Default:     "fallback",
				Description: "An optional variable",
			},
		},
	}

	out, err := Generate(s, Options{Format: "json"})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(out, `"version": "1.0"`) {
		t.Error("expected version in JSON")
	}
	if !strings.Contains(out, `"REQUIRED_VAR"`) {
		t.Error("expected REQUIRED_VAR in JSON")
	}
	if !strings.Contains(out, `"OPTIONAL_VAR"`) {
		t.Error("expected OPTIONAL_VAR in JSON")
	}
	if !strings.Contains(out, `"required": true`) {
		t.Error("expected required field in JSON")
	}
	if !strings.Contains(out, `"default"`) {
		t.Error("expected default field in JSON")
	}
	if !strings.Contains(out, `"description"`) {
		t.Error("expected description field in JSON")
	}
}

func TestGenerateUnsupportedFormat(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"FOO": {Type: schema.TypeString},
		},
	}

	_, err := Generate(s, Options{Format: "xml"})
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestSortedGroupKeys(t *testing.T) {
	groups := map[string][]string{
		"Z": {"Z1"},
		"A": {"A1", "A2"},
		"M": {"M1"},
	}

	keys := sortedGroupKeys(groups)
	expected := []string{"A", "M", "Z"}
	if len(keys) != len(expected) {
		t.Fatalf("expected %d keys, got %d", len(expected), len(keys))
	}
	for i, k := range keys {
		if k != expected[i] {
			t.Errorf("key[%d] = %q, want %q", i, k, expected[i])
		}
	}
}

func TestGenerateMarkdownNoDescription(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"NODESCR": {Type: schema.TypeString},
		},
	}

	out, err := Generate(s, Options{Format: "markdown"})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(out, "NODESCR") {
		t.Error("expected NODESCR in output")
	}
}

func TestGenerateHTMLWithTableAndEmptyLines(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"VAR1": {Type: schema.TypeString, Required: true},
		},
	}

	out, err := Generate(s, Options{Format: "html"})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Should contain table tags
	if !strings.Contains(out, "<table>") {
		t.Error("expected table tag in HTML")
	}
	if !strings.Contains(out, "</table>") {
		t.Error("expected closing table tag in HTML")
	}
}

func TestGenerateHTMLWithOddBackticks(t *testing.T) {
	// Use a variable name with backticks to trigger odd <code> count path
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"VAR_WITH_BACKTICK": {Type: schema.TypeString, Default: "`half"},
		},
	}

	out, err := Generate(s, Options{Format: "html"})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(out, "</html>") {
		t.Error("expected closing html tag")
	}
}
