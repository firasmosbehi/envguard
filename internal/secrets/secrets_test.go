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
			for i, wantRule := range tt.wantRules {
				if i >= len(matches) {
					t.Errorf("missing match for rule %s", wantRule)
					continue
				}
				if matches[i].Rule != wantRule {
					t.Errorf("match[%d].Rule = %q, want %q", i, matches[i].Rule, wantRule)
				}
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
