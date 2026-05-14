// Package secrets provides detection of hardcoded credentials in .env files.
package secrets

import (
	"math"
	"regexp"
)

// Severity represents the severity of a secret finding.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
)

// SecretMatch represents a detected secret in an environment variable.
type SecretMatch struct {
	Key      string   `json:"key"`
	Rule     string   `json:"rule"`
	Message  string   `json:"message"`
	Redacted string   `json:"redacted"`
	Severity Severity `json:"severity"`
}

// Rule defines a single secret detection pattern.
type Rule struct {
	Name       string
	Pattern    *regexp.Regexp
	Message    string
	Severity   Severity
	RedactFunc func(value string) string
}

// Scanner holds secret detection rules.
type Scanner struct {
	rules []Rule
}

// NewScanner creates a scanner with built-in rules plus optional custom rules.
func NewScanner(custom []Rule) *Scanner {
	s := DefaultScanner()
	s.rules = append(s.rules, custom...)
	return s
}

// DefaultScanner returns a scanner with built-in secret detection rules.
func DefaultScanner() *Scanner {
	return &Scanner{
		rules: []Rule{
			{
				Name:     "aws-access-key",
				Pattern:  regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
				Message:  "AWS Access Key ID detected",
				Severity: SeverityHigh,
				RedactFunc: func(_ string) string {
					return "AKIA..."
				},
			},
			{
				Name:     "aws-secret-key",
				Pattern:  regexp.MustCompile(`^[A-Za-z0-9/+=]{40}$`),
				Message:  "AWS Secret Access Key pattern detected (40-character base64-like string)",
				Severity: SeverityCritical,
				RedactFunc: func(v string) string {
					return v[:4] + "..." + v[len(v)-4:]
				},
			},
			{
				Name:     "github-token",
				Pattern:  regexp.MustCompile(`gh[pousr]_[A-Za-z0-9_]{36,}`),
				Message:  "GitHub personal access token detected",
				Severity: SeverityHigh,
				RedactFunc: func(_ string) string {
					return "ghp_..."
				},
			},
			{
				Name:     "private-key",
				Pattern:  regexp.MustCompile(`-----BEGIN (RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`),
				Message:  "Private key detected",
				Severity: SeverityCritical,
				RedactFunc: func(_ string) string {
					return "-----BEGIN PRIVATE KEY----- ... -----END PRIVATE KEY-----"
				},
			},
			{
				Name:     "generic-api-key",
				Pattern:  regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[:=]\s*['"]?([a-z0-9_\-]{16,})['"]?`),
				Message:  "Generic API key pattern detected",
				Severity: SeverityMedium,
				RedactFunc: func(_ string) string {
					return "api_key=..."
				},
			},
			{
				Name:     "slack-token",
				Pattern:  regexp.MustCompile(`xox[baprs]-[0-9]{10,13}-[0-9]{10,13}(-[a-zA-Z0-9]{24})?`),
				Message:  "Slack token detected",
				Severity: SeverityHigh,
				RedactFunc: func(_ string) string {
					return "xoxb-..."
				},
			},
			{
				Name:     "stripe-key",
				Pattern:  regexp.MustCompile(`sk_(live|test)_[0-9a-zA-Z_]{24,}`),
				Message:  "Stripe API key detected",
				Severity: SeverityCritical,
				RedactFunc: func(_ string) string {
					return "sk_live_..."
				},
			},
			{
				Name:     "jwt-token",
				Pattern:  regexp.MustCompile(`eyJ[a-zA-Z0-9_-]*\.eyJ[a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]*`),
				Message:  "JWT token detected",
				Severity: SeverityMedium,
				RedactFunc: func(_ string) string {
					return "eyJ..."
				},
			},
			// New v2.0.0 rules
			{
				Name:     "azure-key",
				Pattern:  regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`),
				Message:  "Azure GUID/API key pattern detected",
				Severity: SeverityMedium,
				RedactFunc: func(v string) string {
					return v[:8] + "..." + v[len(v)-8:]
				},
			},
			{
				Name:     "gcp-api-key",
				Pattern:  regexp.MustCompile(`AIza[0-9A-Za-z_-]{35}`),
				Message:  "Google Cloud Platform API key detected",
				Severity: SeverityHigh,
				RedactFunc: func(_ string) string {
					return "AIza..."
				},
			},
			{
				Name:     "telegram-bot-token",
				Pattern:  regexp.MustCompile(`[0-9]+:AA[0-9A-Za-z_-]{32,}`),
				Message:  "Telegram bot token detected",
				Severity: SeverityHigh,
				RedactFunc: func(v string) string {
					parts := regexp.MustCompile(`:`).Split(v, 2)
					if len(parts) == 2 {
						return parts[0] + ":AA..."
					}
					return "..."
				},
			},
			{
				Name:     "sendgrid-api-key",
				Pattern:  regexp.MustCompile(`SG\.[0-9A-Za-z_-]{20,24}\.[0-9A-Za-z_-]{40,50}`),
				Message:  "SendGrid API key detected",
				Severity: SeverityHigh,
				RedactFunc: func(_ string) string {
					return "SG.xxx..."
				},
			},
			{
				Name:     "twilio-api-key",
				Pattern:  regexp.MustCompile(`SK[0-9a-f]{31,32}`),
				Message:  "Twilio API key detected",
				Severity: SeverityHigh,
				RedactFunc: func(_ string) string {
					return "SK..."
				},
			},
			{
				Name:     "npm-token",
				Pattern:  regexp.MustCompile(`npm_[0-9A-Za-z]{36}`),
				Message:  "npm access token detected",
				Severity: SeverityHigh,
				RedactFunc: func(_ string) string {
					return "npm_..."
				},
			},
			{
				Name:     "docker-config-auth",
				Pattern:  regexp.MustCompile(`"auth"\s*:\s*"[A-Za-z0-9+/=]+"`),
				Message:  "Docker config auth token detected",
				Severity: SeverityMedium,
				RedactFunc: func(_ string) string {
					return `"auth": "..."`
				},
			},
			{
				Name:     "firebase-api-key",
				Pattern:  regexp.MustCompile(`AIza[0-9A-Za-z_-]{35}`),
				Message:  "Firebase API key detected",
				Severity: SeverityMedium,
				RedactFunc: func(_ string) string {
					return "AIza..."
				},
			},
			{
				Name:     "anthropic-api-key",
				Pattern:  regexp.MustCompile(`sk-ant-api[0-9A-Za-z_-]{32,}`),
				Message:  "Anthropic API key detected",
				Severity: SeverityCritical,
				RedactFunc: func(_ string) string {
					return "sk-ant-api..."
				},
			},
			{
				Name:     "openai-api-key",
				Pattern:  regexp.MustCompile(`sk-(proj-|svcacct-)[0-9A-Za-z_-]{40,}`),
				Message:  "OpenAI API key detected",
				Severity: SeverityCritical,
				RedactFunc: func(_ string) string {
					return "sk-..."
				},
			},
		},
	}
}

// Scan checks all values in envVars for secrets.
func (s *Scanner) Scan(envVars map[string]string) []SecretMatch {
	var matches []SecretMatch
	for key, value := range envVars {
		for _, rule := range s.rules {
			if rule.Pattern.MatchString(value) {
				matches = append(matches, SecretMatch{
					Key:      key,
					Rule:     rule.Name,
					Message:  rule.Message,
					Redacted: rule.RedactFunc(value),
					Severity: rule.Severity,
				})
				// Only report first match per rule per key
				break
			}
		}
		// Entropy-based detection for high-entropy strings that look like secrets
		if entropy := shannonEntropy(value); entropy > 4.5 && len(value) >= 20 {
			// Skip if already matched by a rule
			alreadyMatched := false
			for _, m := range matches {
				if m.Key == key {
					alreadyMatched = true
					break
				}
			}
			if !alreadyMatched && !isCommonNonSecret(value) {
				matches = append(matches, SecretMatch{
					Key:      key,
					Rule:     "high-entropy",
					Message:  "High-entropy string detected (possible secret)",
					Redacted: redactHighEntropy(value),
					Severity: SeverityLow,
				})
			}
		}
	}
	return matches
}

// shannonEntropy calculates the Shannon entropy of a string.
// Higher entropy indicates more randomness, which is common in secrets.
func shannonEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}
	freq := make(map[rune]int)
	for _, r := range s {
		freq[r]++
	}
	var entropy float64
	length := float64(len(s))
	for _, count := range freq {
		p := float64(count) / length
		entropy -= p * math.Log2(p)
	}
	return entropy
}

// isCommonNonSecret returns true if the value is likely not a secret.
func isCommonNonSecret(value string) bool {
	// Common non-secret patterns
	commonPatterns := []*regexp.Regexp{
		regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`),                              // numbers
		regexp.MustCompile(`^true|false$`),                                     // booleans
		regexp.MustCompile(`^https?://`),                                       // URLs
		regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`), // emails
		regexp.MustCompile(`^v?[0-9]+\.[0-9]+\.[0-9]+`),                        // semver
		regexp.MustCompile(`^localhost(:[0-9]+)?$`),                            // localhost
		regexp.MustCompile(`^/[a-zA-Z0-9_/\-\.]+$`),                            // file paths
	}
	for _, p := range commonPatterns {
		if p.MatchString(value) {
			return true
		}
	}
	return false
}

// redactHighEntropy redacts a high-entropy string.
func redactHighEntropy(value string) string {
	if len(value) <= 8 {
		return "***"
	}
	return value[:4] + "..." + value[len(value)-4:]
}

// SeverityRank returns a numeric rank for severity comparison.
// Higher number = more severe.
func SeverityRank(s Severity) int {
	switch s {
	case SeverityCritical:
		return 4
	case SeverityHigh:
		return 3
	case SeverityMedium:
		return 2
	case SeverityLow:
		return 1
	default:
		return 0
	}
}

// FilterBySeverity returns only matches with severity >= minSeverity.
func FilterBySeverity(matches []SecretMatch, minSeverity Severity) []SecretMatch {
	minRank := SeverityRank(minSeverity)
	var filtered []SecretMatch
	for _, m := range matches {
		if SeverityRank(m.Severity) >= minRank {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

func isBase64Like(s string) bool {
	for _, c := range s {
		//nolint:staticcheck // Original form is more readable than De Morgan's equivalent.
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '+' || c == '/' || c == '=') {
			return false
		}
	}
	return true
}
