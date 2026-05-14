package secrets

import (
	"testing"
)

func TestScanner(t *testing.T) {
	scanner := DefaultScanner()

	tests := []struct {
		name      string
		envVars   map[string]string
		wantFound bool
		wantRules []string
	}{
		{
			name:      "AWS access key",
			envVars:   map[string]string{"AWS_KEY": "AKIAIOSFODNN7EXAMPLE"},
			wantFound: true,
			wantRules: []string{"aws-access-key"},
		},
		{
			name:      "GitHub token",
			envVars:   map[string]string{"GITHUB_TOKEN": "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"},
			wantFound: true,
			wantRules: []string{"github-token"},
		},
		{
			name:      "private key",
			envVars:   map[string]string{"SSH_KEY": "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA..."},
			wantFound: true,
			wantRules: []string{"private-key"},
		},
		{
			name:      "Stripe key",
			envVars:   map[string]string{"STRIPE_KEY": "sk_test_not_real_pattern_12345678"},
			wantFound: true,
			wantRules: []string{"stripe-key"},
		},
		{
			name:      "JWT token",
			envVars:   map[string]string{"AUTH_TOKEN": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"},
			wantFound: true,
			wantRules: []string{"jwt-token"},
		},
		{
			name:      "no secrets",
			envVars:   map[string]string{"FOO": "bar", "PORT": "3000"},
			wantFound: false,
		},
		{
			name:      "multiple secrets",
			envVars:   map[string]string{"AWS_KEY": "AKIAIOSFODNN7EXAMPLE", "GITHUB_TOKEN": "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"},
			wantFound: true,
			wantRules: []string{"aws-access-key", "github-token"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := scanner.Scan(tt.envVars)
			found := len(matches) > 0
			if found != tt.wantFound {
				t.Errorf("found = %v, want %v; matches = %v", found, tt.wantFound, matches)
			}
			foundRules := make(map[string]bool)
			for _, m := range matches {
				foundRules[m.Rule] = true
			}
			for _, wantRule := range tt.wantRules {
				if !foundRules[wantRule] {
					t.Errorf("missing match for rule %s", wantRule)
				}
			}
		})
	}
}

func TestScanner_NewRules(t *testing.T) {
	scanner := DefaultScanner()

	tests := []struct {
		name     string
		envVars  map[string]string
		wantRule string
	}{
		{
			name:     "Azure GUID",
			envVars:  map[string]string{"AZURE_ID": "550e8400-e29b-41d4-a716-446655440000"},
			wantRule: "azure-key",
		},
		{
			name:     "GCP API key",
			envVars:  map[string]string{"GCP_KEY": "AIzaSyDdI0hCZtE6vySjMm-WEfRq3CPzqKqqsHI"},
			wantRule: "gcp-api-key",
		},
		{
			name:     "Telegram bot token",
			envVars:  map[string]string{"TELEGRAM_TOKEN": "123456789:AAHdqTcvCH1vGWJxfSeofSAs0K5PALDsaw"},
			wantRule: "telegram-bot-token",
		},
		{
			name:     "SendGrid API key",
			envVars:  map[string]string{"SENDGRID_KEY": "SG.xxxxxxxxxxxxxxxxxxxx.xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"},
			wantRule: "sendgrid-api-key",
		},
		{
			name:     "Twilio API key",
			envVars:  map[string]string{"TWILIO_KEY": "SK0123456789abcdef0123456789abcde"},
			wantRule: "twilio-api-key",
		},
		{
			name:     "npm token",
			envVars:  map[string]string{"NPM_TOKEN": "npm_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"},
			wantRule: "npm-token",
		},
		{
			name:     "OpenAI API key",
			envVars:  map[string]string{"OPENAI_KEY": "sk-proj-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"},
			wantRule: "openai-api-key",
		},
		{
			name:     "Anthropic API key",
			envVars:  map[string]string{"ANTHROPIC_KEY": "sk-ant-api03-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"},
			wantRule: "anthropic-api-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := scanner.Scan(tt.envVars)
			found := false
			for _, m := range matches {
				if m.Rule == tt.wantRule {
					found = true
					if m.Severity == "" {
						t.Errorf("expected severity for rule %q to be set", tt.wantRule)
					}
					break
				}
			}
			if !found {
				t.Errorf("expected rule %q to be found, got matches: %v", tt.wantRule, matches)
			}
		})
	}
}

func TestScanner_Severity(t *testing.T) {
	scanner := DefaultScanner()

	tests := []struct {
		name         string
		envVars      map[string]string
		wantRule     string
		wantSeverity Severity
	}{
		{
			name:         "AWS access key is high severity",
			envVars:      map[string]string{"AWS_KEY": "AKIAIOSFODNN7EXAMPLE"},
			wantRule:     "aws-access-key",
			wantSeverity: SeverityHigh,
		},
		{
			name:         "AWS secret key is critical severity",
			envVars:      map[string]string{"AWS_SECRET": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"},
			wantRule:     "aws-secret-key",
			wantSeverity: SeverityCritical,
		},
		{
			name:         "Stripe key is critical severity",
			envVars:      map[string]string{"STRIPE_KEY": "sk_test_not_real_pattern_12345678"},
			wantRule:     "stripe-key",
			wantSeverity: SeverityCritical,
		},
		{
			name:         "JWT token is medium severity",
			envVars:      map[string]string{"AUTH_TOKEN": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"},
			wantRule:     "jwt-token",
			wantSeverity: SeverityMedium,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := scanner.Scan(tt.envVars)
			for _, m := range matches {
				if m.Rule == tt.wantRule {
					if m.Severity != tt.wantSeverity {
						t.Errorf("severity = %q, want %q", m.Severity, tt.wantSeverity)
					}
					return
				}
			}
			t.Errorf("rule %q not found in matches", tt.wantRule)
		})
	}
}

func TestFilterBySeverity(t *testing.T) {
	matches := []SecretMatch{
		{Key: "A", Rule: "r1", Severity: SeverityCritical},
		{Key: "B", Rule: "r2", Severity: SeverityHigh},
		{Key: "C", Rule: "r3", Severity: SeverityMedium},
		{Key: "D", Rule: "r4", Severity: SeverityLow},
	}

	tests := []struct {
		minSev   Severity
		wantLen  int
		wantKeys []string
	}{
		{SeverityCritical, 1, []string{"A"}},
		{SeverityHigh, 2, []string{"A", "B"}},
		{SeverityMedium, 3, []string{"A", "B", "C"}},
		{SeverityLow, 4, []string{"A", "B", "C", "D"}},
	}

	for _, tt := range tests {
		t.Run(string(tt.minSev), func(t *testing.T) {
			filtered := FilterBySeverity(matches, tt.minSev)
			if len(filtered) != tt.wantLen {
				t.Errorf("len = %d, want %d", len(filtered), tt.wantLen)
			}
			for i, wantKey := range tt.wantKeys {
				if i >= len(filtered) || filtered[i].Key != wantKey {
					t.Errorf("filtered[%d].Key = %q, want %q", i, filtered[i].Key, wantKey)
				}
			}
		})
	}
}

func TestShannonEntropy(t *testing.T) {
	tests := []struct {
		input string
		min   float64
		max   float64
	}{
		{"", 0, 0},
		{"aaaaaaaa", 0, 0.1},          // very low entropy
		{"abcdefgh", 2.5, 3.5},        // moderate entropy
		{"aX9#mK2@vL7$qW4", 3.5, 4.5}, // higher entropy
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			entropy := shannonEntropy(tt.input)
			if entropy < tt.min || entropy > tt.max {
				t.Errorf("entropy = %f, want between %f and %f", entropy, tt.min, tt.max)
			}
		})
	}
}

func TestIsCommonNonSecret(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"3000", true},
		{"true", true},
		{"https://example.com", true},
		{"user@example.com", true},
		{"1.2.3", true},
		{"localhost:8080", true},
		{"/path/to/file", true},
		{"AKIAIOSFODNN7EXAMPLE", false},
		{"ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isCommonNonSecret(tt.input)
			if got != tt.want {
				t.Errorf("isCommonNonSecret(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestScanner_EntropyDetection(t *testing.T) {
	scanner := DefaultScanner()

	// High-entropy string that doesn't match any rule
	highEntropy := "aB3xK9mP2vL7qW4nR6tY8uI0oZ1cE5fG2hJ4kL6"
	envVars := map[string]string{"MY_SECRET": highEntropy}

	matches := scanner.Scan(envVars)
	found := false
	for _, m := range matches {
		if m.Rule == "high-entropy" {
			found = true
			if m.Severity != SeverityLow {
				t.Errorf("expected severity low, got %q", m.Severity)
			}
			break
		}
	}
	if !found {
		t.Errorf("expected high-entropy detection for %q", highEntropy)
	}
}

func TestScanner_EntropySkipsCommon(t *testing.T) {
	scanner := DefaultScanner()

	// Common non-secret values should not trigger entropy detection
	envVars := map[string]string{
		"PORT":    "3000",
		"HOST":    "localhost",
		"URL":     "https://example.com",
		"VERSION": "1.2.3",
	}

	matches := scanner.Scan(envVars)
	for _, m := range matches {
		if m.Rule == "high-entropy" {
			t.Errorf("high-entropy should not trigger for common values, got match: %+v", m)
		}
	}
}

func TestSeverityRank(t *testing.T) {
	tests := []struct {
		sev  Severity
		want int
	}{
		{SeverityCritical, 4},
		{SeverityHigh, 3},
		{SeverityMedium, 2},
		{SeverityLow, 1},
		{"unknown", 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.sev), func(t *testing.T) {
			got := SeverityRank(tt.sev)
			if got != tt.want {
				t.Errorf("SeverityRank(%q) = %d, want %d", tt.sev, got, tt.want)
			}
		})
	}
}

func TestIsBase64Like(t *testing.T) {
	if !isBase64Like("abcd1234+/=") {
		t.Error("expected base64-like string to be recognized")
	}
	if isBase64Like("hello world!") {
		t.Error("expected non-base64 string to be rejected")
	}
}
