package schema

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseValidSchema(t *testing.T) {
	yaml := `
version: "1.0"
env:
  DATABASE_URL:
    type: string
    required: true
    description: "PostgreSQL connection string"
  PORT:
    type: integer
    default: 3000
  DEBUG:
    type: boolean
    default: false
  LOG_LEVEL:
    type: string
    enum: [debug, info, warn, error]
    default: info
  API_KEY:
    type: string
    required: true
    pattern: "^[A-Za-z0-9_-]{32,}$"
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "envguard.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	s, err := Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", s.Version)
	}

	if len(s.Env) != 5 {
		t.Errorf("expected 5 variables, got %d", len(s.Env))
	}

	if s.Env["DATABASE_URL"].Type != TypeString || !s.Env["DATABASE_URL"].Required {
		t.Error("DATABASE_URL schema mismatch")
	}

	if s.Env["PORT"].Type != TypeInteger {
		t.Error("PORT type mismatch")
	}

	if s.Env["DEBUG"].Type != TypeBoolean {
		t.Error("DEBUG type mismatch")
	}

	if len(s.Env["LOG_LEVEL"].Enum) != 4 {
		t.Errorf("expected 4 enum values, got %d", len(s.Env["LOG_LEVEL"].Enum))
	}

	if s.Env["API_KEY"].Pattern == "" {
		t.Error("API_KEY pattern should not be empty")
	}
}

func TestParseMissingFile(t *testing.T) {
	_, err := Parse("/nonexistent/envguard.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestParseInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "bad.yaml")
	os.WriteFile(path, []byte("not: [ valid yaml :::"), 0644)

	_, err := Parse(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestValidateSchema(t *testing.T) {
	tests := []struct {
		name    string
		schema  Schema
		wantErr bool
	}{
		{
			name: "valid minimal",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"FOO": {Type: TypeString},
				},
			},
			wantErr: false,
		},
		{
			name: "missing version",
			schema: Schema{
				Env: map[string]*Variable{
					"FOO": {Type: TypeString},
				},
			},
			wantErr: true,
		},
		{
			name: "empty env",
			schema: Schema{
				Version: "1.0",
				Env:     map[string]*Variable{},
			},
			wantErr: true,
		},
		{
			name: "unsupported type",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"FOO": {Type: "datetime"},
				},
			},
			wantErr: true,
		},
		{
			name: "required and default",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"FOO": {Type: TypeString, Required: true, Default: "bar"},
				},
			},
			wantErr: true,
		},
		{
			name: "pattern on non-string",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"FOO": {Type: TypeInteger, Pattern: "^[0-9]+$"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid pattern",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"FOO": {Type: TypeString, Pattern: "[invalid"},
				},
			},
			wantErr: true,
		},
		{
			name: "enum on boolean",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"FOO": {Type: TypeBoolean, Enum: []any{true}},
				},
			},
			wantErr: true,
		},
		{
			name: "valid integer default",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"PORT": {Type: TypeInteger, Default: 3000},
				},
			},
			wantErr: false,
		},
		{
			name: "valid float default",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"RATIO": {Type: TypeFloat, Default: 3.14},
				},
			},
			wantErr: false,
		},
		{
			name: "valid boolean default",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"DEBUG": {Type: TypeBoolean, Default: false},
				},
			},
			wantErr: false,
		},
		{
			name: "nil variable definition",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"FOO": nil,
				},
			},
			wantErr: true,
		},
		{
			name: "valid string enum",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"LEVEL": {Type: TypeString, Enum: []any{"debug", "info"}},
				},
			},
			wantErr: false,
		},
		{
			name: "valid integer enum",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"COUNT": {Type: TypeInteger, Enum: []any{1, 2, 3}},
				},
			},
			wantErr: false,
		},
		{
			name: "valid float enum",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"RATIO": {Type: TypeFloat, Enum: []any{1.5, 2.5}},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid string enum value",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"LEVEL": {Type: TypeString, Enum: []any{123}},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid integer enum value",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"COUNT": {Type: TypeInteger, Enum: []any{"one"}},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid float enum value",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"RATIO": {Type: TypeFloat, Enum: []any{"one"}},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid integer default type",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"COUNT": {Type: TypeInteger, Default: "one"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid float default type",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"RATIO": {Type: TypeFloat, Default: "one"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid boolean default type",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"DEBUG": {Type: TypeBoolean, Default: "true"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid string default type",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"NAME": {Type: TypeString, Default: 123},
				},
			},
			wantErr: true,
		},
		{
			name: "valid integer default as float64",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"COUNT": {Type: TypeInteger, Default: float64(42)},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid integer default as non-whole float64",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"COUNT": {Type: TypeInteger, Default: float64(3.14)},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.schema.Validate()
			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestIsEnvVarNameValid(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{"FOO", true},
		{"FOO_BAR", true},
		{"FOO_BAR_123", true},
		{"123_FOO", false},
		{"", false},
		{"FOO-BAR", false},
		{"FOO.BAR", false},
		{"FOO BAR", false},
	}

	for _, tt := range tests {
		got := IsEnvVarNameValid(tt.name)
		if got != tt.valid {
			t.Errorf("IsEnvVarNameValid(%q) = %v, want %v", tt.name, got, tt.valid)
		}
	}
}

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"FOO", "FOO"},
		{"  FOO  ", "FOO"},
		{"", ""},
	}

	for _, tt := range tests {
		got := NormalizeName(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
