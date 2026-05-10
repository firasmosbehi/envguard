package schema

import (
	"testing"
)

func TestValidateSchemaMinMax(t *testing.T) {
	tests := []struct {
		name    string
		schema  Schema
		wantErr bool
	}{
		{
			name: "valid integer min",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"PORT": {Type: TypeInteger, Min: 1024},
				},
			},
			wantErr: false,
		},
		{
			name: "valid integer max",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"PORT": {Type: TypeInteger, Max: 65535},
				},
			},
			wantErr: false,
		},
		{
			name: "valid integer min and max",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"PORT": {Type: TypeInteger, Min: 1024, Max: 65535},
				},
			},
			wantErr: false,
		},
		{
			name: "valid float min",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"RATIO": {Type: TypeFloat, Min: 0.0},
				},
			},
			wantErr: false,
		},
		{
			name: "min greater than max",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"PORT": {Type: TypeInteger, Min: 1000, Max: 100},
				},
			},
			wantErr: true,
		},
		{
			name: "min on string type",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"FOO": {Type: TypeString, Min: 1},
				},
			},
			wantErr: true,
		},
		{
			name: "max on boolean type",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"FOO": {Type: TypeBoolean, Max: 1},
				},
			},
			wantErr: true,
		},
		{
			name: "non-numeric min",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"FOO": {Type: TypeInteger, Min: "abc"},
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

func TestValidateSchemaStringLength(t *testing.T) {
	min5 := 5
	max10 := 10
	max3 := 3

	tests := []struct {
		name    string
		schema  Schema
		wantErr bool
	}{
		{
			name: "valid minLength",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"FOO": {Type: TypeString, MinLength: &min5},
				},
			},
			wantErr: false,
		},
		{
			name: "valid maxLength",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"FOO": {Type: TypeString, MaxLength: &max10},
				},
			},
			wantErr: false,
		},
		{
			name: "valid minLength and maxLength",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"FOO": {Type: TypeString, MinLength: &min5, MaxLength: &max10},
				},
			},
			wantErr: false,
		},
		{
			name: "minLength greater than maxLength",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"FOO": {Type: TypeString, MinLength: &min5, MaxLength: &max3},
				},
			},
			wantErr: true,
		},
		{
			name: "minLength on integer",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"FOO": {Type: TypeInteger, MinLength: &min5},
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

func TestValidateSchemaFormat(t *testing.T) {
	tests := []struct {
		name    string
		schema  Schema
		wantErr bool
	}{
		{
			name: "valid email format",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"EMAIL": {Type: TypeString, Format: "email"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid url format",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"ENDPOINT": {Type: TypeString, Format: "url"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid uuid format",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"ID": {Type: TypeString, Format: "uuid"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid format",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"FOO": {Type: TypeString, Format: "phone"},
				},
			},
			wantErr: true,
		},
		{
			name: "format on integer",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"FOO": {Type: TypeInteger, Format: "email"},
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

func TestValidateSchemaDisallow(t *testing.T) {
	tests := []struct {
		name    string
		schema  Schema
		wantErr bool
	}{
		{
			name: "valid disallow",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"API_KEY": {Type: TypeString, Disallow: []string{"undefined", "null", ""}},
				},
			},
			wantErr: false,
		},
		{
			name: "disallow on integer",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"FOO": {Type: TypeInteger, Disallow: []string{"abc"}},
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

func TestValidateSchemaEnvironmentRules(t *testing.T) {
	tests := []struct {
		name    string
		schema  Schema
		wantErr bool
	}{
		{
			name: "valid requiredIn",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"DB_URL": {Type: TypeString, RequiredIn: []string{"production", "staging"}},
				},
			},
			wantErr: false,
		},
		{
			name: "valid devOnly",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"DEBUG_TOOL": {Type: TypeString, DevOnly: true},
				},
			},
			wantErr: false,
		},
		{
			name: "devOnly and required conflict",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"FOO": {Type: TypeString, DevOnly: true, Required: true},
				},
			},
			wantErr: true,
		},
		{
			name: "devOnly and requiredIn conflict",
			schema: Schema{
				Version: "1.0",
				Env: map[string]*Variable{
					"FOO": {Type: TypeString, DevOnly: true, RequiredIn: []string{"production"}},
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
