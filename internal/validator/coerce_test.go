package validator

import (
	"testing"
)

func TestCoerceString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"  hello  ", "hello"},
		{"", ""},
	}

	for _, tt := range tests {
		got, err := coerceString(tt.input)
		if err != nil {
			t.Errorf("coerceString(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.expected {
			t.Errorf("coerceString(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestCoerceInteger(t *testing.T) {
	tests := []struct {
		input   string
		want    int64
		wantErr bool
	}{
		{"42", 42, false},
		{"-3", -3, false},
		{"0", 0, false},
		{"  123  ", 123, false},
		{"3.14", 0, true},
		{"abc", 0, true},
		{"12.0", 0, true},
		{"", 0, true},
		{"true", 0, true},
	}

	for _, tt := range tests {
		got, err := coerceInteger(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("coerceInteger(%q) expected error", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("coerceInteger(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("coerceInteger(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestCoerceFloat(t *testing.T) {
	tests := []struct {
		input   string
		want    float64
		wantErr bool
	}{
		{"3.14", 3.14, false},
		{"-2.5", -2.5, false},
		{"10", 10, false},
		{"  1.5  ", 1.5, false},
		{"abc", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		got, err := coerceFloat(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("coerceFloat(%q) expected error", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("coerceFloat(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("coerceFloat(%q) = %f, want %f", tt.input, got, tt.want)
		}
	}
}

func TestCoerceBoolean(t *testing.T) {
	tests := []struct {
		input   string
		want    bool
		wantErr bool
	}{
		{"true", true, false},
		{"TRUE", true, false},
		{"True", true, false},
		{"1", true, false},
		{"yes", true, false},
		{"YES", true, false},
		{"on", true, false},
		{"ON", true, false},
		{"false", false, false},
		{"FALSE", false, false},
		{"False", false, false},
		{"0", false, false},
		{"no", false, false},
		{"NO", false, false},
		{"off", false, false},
		{"OFF", false, false},
		{"2", false, true},
		{"maybe", false, true},
		{"", false, true},
		{"  true  ", true, false},
	}

	for _, tt := range tests {
		got, err := coerceBoolean(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("coerceBoolean(%q) expected error", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("coerceBoolean(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("coerceBoolean(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
