package schema

import (
	"os"
	"testing"
)

func TestValidateTransform(t *testing.T) {
	tests := []struct {
		name      string
		transform string
		varType   Type
		wantErr   bool
	}{
		{"valid lowercase on string", "lowercase", TypeString, false},
		{"valid uppercase on string", "uppercase", TypeString, false},
		{"valid trim on string", "trim", TypeString, false},
		{"invalid transform on integer", "lowercase", TypeInteger, true},
		{"unsupported transform", "capitalize", TypeString, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"FOO": {Type: tt.varType, Transform: tt.transform},
				},
			}
			err := s.Validate()
			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateNewFormats(t *testing.T) {
	formats := []string{"duration", "semver", "hostname", "hex", "cron"}
	for _, f := range formats {
		t.Run(f, func(t *testing.T) {
			s := &Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"FOO": {Type: TypeString, Format: f},
				},
			}
			if err := s.Validate(); err != nil {
				t.Errorf("unexpected error for format %q: %v", f, err)
			}
		})
	}
}

func TestValidateCustomSecretRules(t *testing.T) {
	tests := []struct {
		name    string
		secrets *Secrets
		wantErr bool
	}{
		{
			name: "valid custom rule",
			secrets: &Secrets{
				Custom: []CustomSecretRule{
					{Name: "my-token", Pattern: `iat_[a-z0-9]{32}`, Message: "My token detected"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			secrets: &Secrets{
				Custom: []CustomSecretRule{
					{Pattern: `iat_[a-z0-9]{32}`, Message: "Missing name"},
				},
			},
			wantErr: true,
		},
		{
			name: "missing pattern",
			secrets: &Secrets{
				Custom: []CustomSecretRule{
					{Name: "no-pattern", Message: "No pattern"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid pattern",
			secrets: &Secrets{
				Custom: []CustomSecretRule{
					{Name: "bad-pattern", Pattern: `[invalid`, Message: "Bad regex"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"FOO": {Type: TypeString},
				},
				Secrets: tt.secrets,
			}
			err := s.Validate()
			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestParseLenient(t *testing.T) {
	tmpDir := t.TempDir()
	schemaPath := tmpDir + "/envguard.yaml"

	content := `
version: "1.0"
env:
  FOO:
    type: string
    required: true
    default: "bar"
`
	if err := writeFile(schemaPath, content); err != nil {
		t.Fatalf("failed to write schema: %v", err)
	}

	s, err := ParseLenient(schemaPath)
	if err != nil {
		t.Fatalf("ParseLenient failed: %v", err)
	}
	if s.Version != "1.0" {
		t.Errorf("expected version 1.0, got %q", s.Version)
	}
	if s.Env["FOO"].Required != true {
		t.Errorf("expected FOO to be required")
	}
	if s.Env["FOO"].Default != "bar" {
		t.Errorf("expected FOO default to be bar, got %v", s.Env["FOO"].Default)
	}
}

func writeFile(path, content string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}
