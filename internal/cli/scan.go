package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/envguard/envguard/internal/dotenv"
	"github.com/envguard/envguard/internal/reporter"
	"github.com/envguard/envguard/internal/schema"
	"github.com/envguard/envguard/internal/secrets"
)

type scanOptions struct {
	envPaths       []string
	format         string
	schemaPath     string
	secretSeverity string
	baselinePath   string
}

type baselineEntry struct {
	Key      string           `json:"key"`
	Rule     string           `json:"rule"`
	Severity secrets.Severity `json:"severity"`
}

func newScanCmd() *cobra.Command {
	opts := &scanOptions{}

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan .env files for hardcoded secrets",
		Long:  `Scan checks .env values for patterns that match known secret types (API keys, tokens, private keys, etc.).`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runScan(cmd.OutOrStdout(), cmd.ErrOrStderr(), opts)
		},
		SilenceUsage: true,
	}

	cmd.Flags().StringArrayVarP(&opts.envPaths, "env", "e", []string{".env"}, "Path to .env file (can be specified multiple times)")
	cmd.Flags().StringVarP(&opts.format, "format", "f", "text", "Output format: text, json, or sarif")
	cmd.Flags().StringVarP(&opts.schemaPath, "schema", "s", "", "Optional schema file with custom secret rules")
	cmd.Flags().StringVar(&opts.secretSeverity, "secret-severity", "low", "Minimum severity to report: critical, high, medium, low")
	cmd.Flags().StringVar(&opts.baselinePath, "baseline", "", "Path to baseline file to ignore known findings")

	return cmd
}

func runScan(stdout, stderr io.Writer, opts *scanOptions) error {
	// Parse .env files
	envVars := make(map[string]string)
	for _, path := range opts.envPaths {
		vars, err := dotenv.Parse(path)
		if err != nil {
			fmt.Fprintf(stderr, "Error: %v\n", err)
			return ErrIO
		}
		for k, v := range vars {
			envVars[k] = v
		}
	}

	var scanner *secrets.Scanner
	if opts.schemaPath != "" {
		s, err := schema.Parse(opts.schemaPath)
		if err != nil {
			fmt.Fprintf(stderr, "Error: %v\n", err)
			return ErrIO
		}
		if s.Secrets != nil && len(s.Secrets.Custom) > 0 {
			customRules := make([]secrets.Rule, 0, len(s.Secrets.Custom))
			for _, cr := range s.Secrets.Custom {
				sev := secrets.SeverityMedium
				if cr.Severity != "" {
					sev = secrets.Severity(cr.Severity)
				}
				customRules = append(customRules, secrets.Rule{
					Name:     cr.Name,
					Pattern:  regexp.MustCompile(cr.Pattern),
					Message:  cr.Message,
					Severity: sev,
					RedactFunc: func(_ string) string {
						return "***"
					},
				})
			}
			scanner = secrets.NewScanner(customRules)
		} else {
			scanner = secrets.DefaultScanner()
		}
	} else {
		scanner = secrets.DefaultScanner()
	}
	matches := scanner.Scan(envVars)

	// Filter by severity
	minSev := secrets.Severity(opts.secretSeverity)
	matches = secrets.FilterBySeverity(matches, minSev)

	// Apply baseline if provided
	if opts.baselinePath != "" {
		baseline, err := loadBaseline(opts.baselinePath)
		if err != nil {
			fmt.Fprintf(stderr, "Error: failed to load baseline: %v\n", err)
			return ErrIO
		}
		matches = filterBaseline(matches, baseline)
	}

	if len(matches) == 0 {
		fmt.Fprintln(stdout, "✓ No secrets detected.")
		return nil
	}

	found := false
	switch opts.format {
	case "json":
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(matches); err != nil {
			fmt.Fprintf(stderr, "Error: failed to format output: %v\n", err)
			return ErrIO
		}
		found = true
	case "sarif":
		if err := reporter.SARIFScan(stdout, matches, opts.envPaths, version); err != nil {
			fmt.Fprintf(stderr, "Error: failed to format output: %v\n", err)
			return ErrIO
		}
		found = true
	case "text":
		fmt.Fprintf(stdout, "✗ Secret scan found %d issue(s)\n\n", len(matches))
		for _, m := range matches {
			sevLabel := ""
			if m.Severity != "" {
				sevLabel = fmt.Sprintf(" [%s]", m.Severity)
			}
			fmt.Fprintf(stdout, "  • %s%s\n", m.Key, sevLabel)
			fmt.Fprintf(stdout, "    └─ %s: %s (redacted: %s)\n", m.Rule, m.Message, m.Redacted)
		}
		found = true
	default:
		fmt.Fprintf(stderr, "Error: unknown format %q\n", opts.format)
		return ErrIO
	}

	if found {
		return ErrValidationFailed
	}

	return nil
}

func loadBaseline(path string) ([]baselineEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entries []baselineEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func filterBaseline(matches []secrets.SecretMatch, baseline []baselineEntry) []secrets.SecretMatch {
	baselineSet := make(map[string]bool)
	for _, b := range baseline {
		baselineSet[b.Key+"::"+b.Rule] = true
	}
	var filtered []secrets.SecretMatch
	for _, m := range matches {
		if !baselineSet[m.Key+"::"+m.Rule] {
			filtered = append(filtered, m)
		}
	}
	return filtered
}
