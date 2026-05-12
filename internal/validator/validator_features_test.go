package validator

import (
	"strings"
	"testing"

	"github.com/envguard/envguard/internal/schema"
)

func TestDeprecatedWarning(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"OLD_VAR": {
				Type:       schema.TypeString,
				Deprecated: "Use NEW_VAR instead",
			},
		},
	}

	envVars := map[string]string{"OLD_VAR": "value"}
	result := Validate(s, envVars, false, "")

	if result.Valid != true {
		t.Errorf("expected valid, got invalid")
	}
	if len(result.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(result.Warnings))
	}
	if !strings.Contains(result.Warnings[0].Message, "deprecated") {
		t.Errorf("expected deprecated warning, got: %s", result.Warnings[0].Message)
	}
}

func TestSensitiveRedaction(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"PASSWORD": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				Pattern:   "^admin$",
			},
		},
	}

	envVars := map[string]string{"PASSWORD": "supersecret123"}
	result := Validate(s, envVars, false, "")
	result.RedactSensitive(envVars, s)

	if result.Valid != false {
		t.Errorf("expected invalid, got valid")
	}
	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}
	if strings.Contains(result.Errors[0].Message, "supersecret123") {
		t.Errorf("expected sensitive value to be redacted, got: %s", result.Errors[0].Message)
	}
	if !strings.Contains(result.Errors[0].Message, "***") {
		t.Errorf("expected *** redaction, got: %s", result.Errors[0].Message)
	}
}

func TestTransformLowercase(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"LOG_LEVEL": {
				Type:      schema.TypeString,
				Transform: "lowercase",
				Enum:      []any{"debug", "info", "warn", "error"},
			},
		},
	}

	envVars := map[string]string{"LOG_LEVEL": "DEBUG"}
	result := Validate(s, envVars, false, "")

	if !result.Valid {
		t.Errorf("expected valid after lowercase transform, got errors: %v", result.Errors)
	}
}

func TestTransformUppercase(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"MODE": {
				Type:      schema.TypeString,
				Transform: "uppercase",
				Enum:      []any{"DEV", "PROD"},
			},
		},
	}

	envVars := map[string]string{"MODE": "dev"}
	result := Validate(s, envVars, false, "")

	if !result.Valid {
		t.Errorf("expected valid after uppercase transform, got errors: %v", result.Errors)
	}
}

func TestTransformTrim(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"NAME": {
				Type:      schema.TypeString,
				Transform: "trim",
				Pattern:   "^hello$",
			},
		},
	}

	envVars := map[string]string{"NAME": "  hello  "}
	result := Validate(s, envVars, false, "")

	if !result.Valid {
		t.Errorf("expected valid after trim transform, got errors: %v", result.Errors)
	}
}

func TestFormatDuration(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"TIMEOUT": {
				Type:   schema.TypeString,
				Format: "duration",
			},
		},
	}

	tests := []struct {
		value     string
		wantValid bool
	}{
		{"5m", true},
		{"2h30m", true},
		{"1s", true},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			result := Validate(s, map[string]string{"TIMEOUT": tt.value}, false, "")
			if result.Valid != tt.wantValid {
				t.Errorf("Validate(%q) valid=%v, want %v", tt.value, result.Valid, tt.wantValid)
			}
		})
	}
}

func TestFormatSemver(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"VERSION": {
				Type:   schema.TypeString,
				Format: "semver",
			},
		},
	}

	tests := []struct {
		value     string
		wantValid bool
	}{
		{"1.2.3", true},
		{"0.0.1", true},
		{"1.2.3-beta.1", true},
		{"1.2.3+build.123", true},
		{"1.2", false},
		{"v1.2.3", false},
		{"not-semver", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			result := Validate(s, map[string]string{"VERSION": tt.value}, false, "")
			if result.Valid != tt.wantValid {
				t.Errorf("Validate(%q) valid=%v, want %v", tt.value, result.Valid, tt.wantValid)
			}
		})
	}
}

func TestFormatHostname(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"HOST": {
				Type:     schema.TypeString,
				Format:   "hostname",
				Required: true,
			},
		},
	}

	tests := []struct {
		value     string
		wantValid bool
	}{
		{"example.com", true},
		{"api.example.com", true},
		{"localhost", true},
		{"-invalid.com", false},
		{"", false},
		{"a.b.c.d.e.f.g.h.i.j.k.l.m.n.o.p.q.r.s.t.u.v.w.x.y.z.example.com.this.is.too.long.for.a.hostname.because.it.exceeds.two.hundred.and.fifty.three.characters.total.length.which.is.the.maximum.allowed.for.a.fully.qualified.domain.name.in.the.dns.system.and.this.string.is.clearly.longer.than.that.limit", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			result := Validate(s, map[string]string{"HOST": tt.value}, false, "")
			if result.Valid != tt.wantValid {
				t.Errorf("Validate(%q) valid=%v, want %v", tt.value, result.Valid, tt.wantValid)
			}
		})
	}
}

func TestFormatHex(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"COLOR": {
				Type:     schema.TypeString,
				Format:   "hex",
				Required: true,
			},
		},
	}

	tests := []struct {
		value     string
		wantValid bool
	}{
		{"FF5733", true},
		{"0xFF5733", true},
		{"deadbeef", true},
		{"0xDEADBEEF", true},
		{"GHIJKL", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			result := Validate(s, map[string]string{"COLOR": tt.value}, false, "")
			if result.Valid != tt.wantValid {
				t.Errorf("Validate(%q) valid=%v, want %v", tt.value, result.Valid, tt.wantValid)
			}
		})
	}
}

func TestFormatCron(t *testing.T) {
	s := &schema.Schema{
		Version: "1.0",
		Env: map[string]*schema.Variable{
			"SCHEDULE": {
				Type:     schema.TypeString,
				Format:   "cron",
				Required: true,
			},
		},
	}

	tests := []struct {
		value     string
		wantValid bool
	}{
		{"0 0 * * *", true},
		{"*/5 * * * *", true},
		{"@daily", true},
		{"@hourly", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			result := Validate(s, map[string]string{"SCHEDULE": tt.value}, false, "")
			if result.Valid != tt.wantValid {
				t.Errorf("Validate(%q) valid=%v, want %v", tt.value, result.Valid, tt.wantValid)
			}
		})
	}
}
