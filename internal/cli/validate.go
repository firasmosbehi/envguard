package cli

import (
	"fmt"
	"io"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/envguard/envguard/internal/dotenv"
	"github.com/envguard/envguard/internal/reporter"
	"github.com/envguard/envguard/internal/schema"
	"github.com/envguard/envguard/internal/secrets"
	"github.com/envguard/envguard/internal/validator"
)

type validateOptions struct {
	schemaPath  string
	envPaths    []string
	format      string
	strict      bool
	envName     string
	scanSecrets bool
}

func newValidateCmd() *cobra.Command {
	opts := &validateOptions{}

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a .env file against a schema",
		Long:  `Validate checks that the given .env file satisfies the rules defined in the schema YAML file.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runValidate(cmd.OutOrStdout(), cmd.ErrOrStderr(), opts)
		},
		SilenceUsage: true,
	}

	cmd.Flags().StringVarP(&opts.schemaPath, "schema", "s", "envguard.yaml", "Path to schema YAML file")
	cmd.Flags().StringArrayVarP(&opts.envPaths, "env", "e", []string{".env"}, "Path to .env file (can be specified multiple times)")
	cmd.Flags().StringVarP(&opts.format, "format", "f", "text", "Output format: text, json, or github")
	cmd.Flags().BoolVar(&opts.strict, "strict", false, "Fail if .env contains keys not defined in schema")
	cmd.Flags().StringVar(&opts.envName, "env-name", "", "Environment name (e.g. production, development) for environment-specific rules")
	cmd.Flags().BoolVar(&opts.scanSecrets, "scan-secrets", false, "Scan for hardcoded secrets in .env values")

	return cmd
}

func runValidate(stdout, stderr io.Writer, opts *validateOptions) error {
	// Parse schema
	s, err := schema.Parse(opts.schemaPath)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return ErrIO
	}

	// Parse .env files (later files override earlier ones)
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

	// Validate
	result := validator.Validate(s, envVars, opts.strict, opts.envName)

	// Redact sensitive values from output
	result.RedactSensitive(envVars, s)

	// Scan for secrets if requested
	if opts.scanSecrets {
		var scanner *secrets.Scanner
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
		secretMatches := scanner.Scan(envVars)
		for _, m := range secretMatches {
			result.AddError(m.Key, "secret", m.Message+" (redacted: "+m.Redacted+")")
		}
	}

	// Report
	switch opts.format {
	case "json":
		if err := reporter.JSON(stdout, result); err != nil {
			fmt.Fprintf(stderr, "Error: failed to format output: %v\n", err)
			return ErrIO
		}
	case "github":
		reporter.GitHub(stdout, result, opts.envPaths)
	case "text":
		reporter.Text(stdout, result)
	default:
		fmt.Fprintf(stderr, "Error: unknown format %q\n", opts.format)
		return ErrIO
	}

	if !result.Valid {
		return ErrValidationFailed
	}

	return nil
}
