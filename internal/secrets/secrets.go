// Package secrets provides detection of hardcoded credentials in .env files.
package secrets

import (
	"regexp"
)

// SecretMatch represents a detected secret in an environment variable.
type SecretMatch struct {
	Key      string `json:"key"`
	Rule     string `json:"rule"`
	Message  string `json:"message"`
	Redacted string `json:"redacted"`
}

// Scanner holds secret detection rules.
type Scanner struct {
	rules []Rule
}

// Rule defines a single secret detection pattern.
type Rule struct {
	Name       string
	Pattern    *regexp.Regexp
	Message    string
	RedactFunc func(value string) string
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
				Name:    "aws-access-key",
				Pattern: regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
				Message: "AWS Access Key ID detected",
				RedactFunc: func(v string) string {
					return redactMatch(v, `AKIA[0-9A-Z]{16}`, "AKIA...")
				},
			},
			{
				Name:    "aws-secret-key",
				Pattern: regexp.MustCompile(`^[A-Za-z0-9/+=]{40}$`),
				Message: "AWS Secret Access Key pattern detected (40-character base64-like string)",
				RedactFunc: func(v string) string {
					return v[:4] + "..." + v[len(v)-4:]
				},
			},
			{
				Name:    "github-token",
				Pattern: regexp.MustCompile(`gh[pousr]_[A-Za-z0-9_]{36,}`),
				Message: "GitHub personal access token detected",
				RedactFunc: func(v string) string {
					return redactMatch(v, `gh[pousr]_[A-Za-z0-9_]{36,}`, "ghp_...")
				},
			},
			{
				Name:    "private-key",
				Pattern: regexp.MustCompile(`-----BEGIN (RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`),
				Message: "Private key detected",
				RedactFunc: func(v string) string {
					return "-----BEGIN PRIVATE KEY----- ... -----END PRIVATE KEY-----"
				},
			},
			{
				Name:    "generic-api-key",
				Pattern: regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[:=]\s*['"]?([a-z0-9_\-]{16,})['"]?`),
				Message: "Generic API key pattern detected",
				RedactFunc: func(v string) string {
					return redactMatch(v, `(?i)(api[_-]?key|apikey)\s*[:=]\s*['"]?([a-z0-9_\-]{16,})['"]?`, "api_key=...")
				},
			},
			{
				Name:    "slack-token",
				Pattern: regexp.MustCompile(`xox[baprs]-[0-9]{10,13}-[0-9]{10,13}(-[a-zA-Z0-9]{24})?`),
				Message: "Slack token detected",
				RedactFunc: func(v string) string {
					return "xoxb-..."
				},
			},
			{
				Name:    "stripe-key",
				Pattern: regexp.MustCompile(`sk_(live|test)_[0-9a-zA-Z_]{24,}`),
				Message: "Stripe API key detected",
				RedactFunc: func(v string) string {
					return "sk_live_..."
				},
			},
			{
				Name:    "jwt-token",
				Pattern: regexp.MustCompile(`eyJ[a-zA-Z0-9_-]*\.eyJ[a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]*`),
				Message: "JWT token detected",
				RedactFunc: func(v string) string {
					return "eyJ..."
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
				})
				// Only report first match per rule per key
				break
			}
		}
	}
	return matches
}

func isBase64Like(s string) bool {
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '+' || c == '/' || c == '=') {
			return false
		}
	}
	return true
}

func redactMatch(value, pattern, replacement string) string {
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(value, replacement)
}
