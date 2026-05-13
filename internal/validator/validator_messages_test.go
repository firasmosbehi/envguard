package validator

import (
	"testing"

	"github.com/envguard/envguard/internal/schema"
)

func TestCustomMessages(t *testing.T) {
	tests := []struct {
		name     string
		variable *schema.Variable
		value    string
		wantMsg  string
		wantRule string
	}{
		{
			name:     "custom required message",
			variable: &schema.Variable{Type: schema.TypeString, Required: true, Message: "API_KEY is required for authentication"},
			value:    "",
			wantMsg:  "API_KEY is required for authentication",
			wantRule: "required",
		},
		{
			name:     "custom min message",
			variable: &schema.Variable{Type: schema.TypeInteger, Min: 1024, Message: "PORT must be a valid port number"},
			value:    "80",
			wantMsg:  "PORT must be a valid port number",
			wantRule: "min",
		},
		{
			name:     "custom pattern message",
			variable: &schema.Variable{Type: schema.TypeString, Pattern: "^[a-z]+$", Message: "Only lowercase letters allowed"},
			value:    "ABC",
			wantMsg:  "Only lowercase letters allowed",
			wantRule: "pattern",
		},
		{
			name:     "custom format message",
			variable: &schema.Variable{Type: schema.TypeString, Format: "email", Message: "Please provide a valid email"},
			value:    "bad",
			wantMsg:  "Please provide a valid email",
			wantRule: "format",
		},
		{
			name:     "custom enum message",
			variable: &schema.Variable{Type: schema.TypeString, Enum: []any{"a", "b"}, Message: "Must be a or b"},
			value:    "c",
			wantMsg:  "Must be a or b",
			wantRule: "enum",
		},
		{
			name:     "default message when no custom",
			variable: &schema.Variable{Type: schema.TypeString, Required: true},
			value:    "",
			wantMsg:  "variable is missing or empty",
			wantRule: "required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &schema.Schema{
				Version: "1.0",
				Env:     map[string]*schema.Variable{"TEST": tt.variable},
			}
			result := Validate(s, map[string]string{"TEST": tt.value}, false, "")
			if result.Valid {
				t.Fatalf("expected validation to fail")
			}
			if len(result.Errors) == 0 {
				t.Fatalf("expected at least one error")
			}
			if result.Errors[0].Rule != tt.wantRule {
				t.Errorf("Rule = %q, want %q", result.Errors[0].Rule, tt.wantRule)
			}
			if result.Errors[0].Message != tt.wantMsg {
				t.Errorf("Message = %q, want %q", result.Errors[0].Message, tt.wantMsg)
			}
		})
	}
}
