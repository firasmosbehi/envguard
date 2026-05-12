package secrets

import (
	"regexp"
	"testing"
)

func TestNewScannerWithCustomRules(t *testing.T) {
	custom := []Rule{
		{
			Name:    "internal-token",
			Pattern: regexp.MustCompile(`iat_[a-zA-Z0-9]{32}`),
			Message: "Internal API token detected",
			RedactFunc: func(v string) string {
				return "***"
			},
		},
	}

	scanner := NewScanner(custom)
	envVars := map[string]string{
		"API_TOKEN": "iat_12345678901234567890123456789012",
	}

	matches := scanner.Scan(envVars)
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if matches[0].Rule != "internal-token" {
		t.Errorf("expected rule 'internal-token', got %q", matches[0].Rule)
	}
}

func TestDefaultScannerDoesNotIncludeCustomRules(t *testing.T) {
	scanner := DefaultScanner()
	envVars := map[string]string{
		"API_TOKEN": "iat_12345678901234567890123456789012",
	}

	matches := scanner.Scan(envVars)
	if len(matches) != 0 {
		t.Errorf("expected 0 matches from default scanner, got %d", len(matches))
	}
}
