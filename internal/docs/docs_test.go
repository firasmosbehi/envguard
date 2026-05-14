package docs

import (
	"strings"
	"testing"

	"github.com/envguard/envguard/internal/schema"
)

func TestGenerateMarkdown(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"API_KEY": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "API key for external service",
				Sensitive:   true,
			},
			"PORT": {
				Type:    schema.TypeInteger,
				Default: 3000,
			},
		},
	}

	out, err := Generate(s, Options{Format: "markdown"})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(out, "API_KEY") {
		t.Error("expected output to contain API_KEY")
	}
	if !strings.Contains(out, "PORT") {
		t.Error("expected output to contain PORT")
	}
	if !strings.Contains(out, "Required") {
		t.Error("expected output to contain Required")
	}
}

func TestGenerateHTML(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"API_KEY": {Type: schema.TypeString, Required: true},
		},
	}

	out, err := Generate(s, Options{Format: "html"})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(out, "<html>") {
		t.Error("expected HTML output")
	}
	if !strings.Contains(out, "API_KEY") {
		t.Error("expected output to contain API_KEY")
	}
}

func TestGenerateJSON(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"API_KEY": {Type: schema.TypeString},
		},
	}

	out, err := Generate(s, Options{Format: "json"})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(out, `"version"`) {
		t.Error("expected JSON to contain version")
	}
}

func TestGroupByPrefix(t *testing.T) {
	groups := groupByPrefix([]string{"DB_HOST", "DB_PORT", "API_KEY", "REDIS_URL"})
	if len(groups["DB"]) != 2 {
		t.Errorf("expected 2 DB vars, got %d", len(groups["DB"]))
	}
	if len(groups["API"]) != 1 {
		t.Errorf("expected 1 API var, got %d", len(groups["API"]))
	}
}
