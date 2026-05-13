package cli

import (
	"fmt"
	"io"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/envguard/envguard/internal/dotenv"
	"github.com/envguard/envguard/internal/schema"
	"github.com/envguard/envguard/internal/secrets"
)

type scanOptions struct {
	envPaths   []string
	format     string
	schemaPath string
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
	cmd.Flags().StringVarP(&opts.format, "format", "f", "text", "Output format: text or json")
	cmd.Flags().StringVarP(&opts.schemaPath, "schema", "s", "", "Optional schema file with custom secret rules")

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
				customRules = append(customRules, secrets.Rule{
					Name:    cr.Name,
					Pattern: regexp.MustCompile(cr.Pattern),
					Message: cr.Message,
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

	if len(matches) == 0 {
		fmt.Fprintln(stdout, "✓ No secrets detected.")
		return nil
	}

	found := false
	switch opts.format {
	case "json":
		fmt.Fprintln(stdout, "[")
		for i, m := range matches {
			comma := ","
			if i == len(matches)-1 {
				comma = ""
			}
			fmt.Fprintf(stdout, `  {"key": %q, "rule": %q, "message": %q, "redacted": %q}%s`+"\n", m.Key, m.Rule, m.Message, m.Redacted, comma)
		}
		fmt.Fprintln(stdout, "]")
		found = true
	case "text":
		fmt.Fprintf(stdout, "✗ Secret scan found %d issue(s)\n\n", len(matches))
		for _, m := range matches {
			fmt.Fprintf(stdout, "  • %s\n", m.Key)
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
