package cli

import (
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/envguard/envguard/internal/config"
	"github.com/envguard/envguard/internal/dotenv"
	"github.com/envguard/envguard/internal/reporter"
	"github.com/envguard/envguard/internal/schema"
	"github.com/envguard/envguard/internal/secrets"
	"github.com/envguard/envguard/internal/validator"
)

type validateOptions struct {
	schemaPath     string
	envPaths       []string
	format         string
	strict         bool
	envName        string
	scanSecrets    bool
	configPath     string
	failOnWarnings bool
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

	cmd.Flags().StringVarP(&opts.schemaPath, "schema", "s", "", "Path to schema YAML file")
	cmd.Flags().StringArrayVarP(&opts.envPaths, "env", "e", nil, "Path to .env file (can be specified multiple times)")
	cmd.Flags().StringVarP(&opts.format, "format", "f", "", "Output format: text, json, github, or sarif")
	cmd.Flags().BoolVar(&opts.strict, "strict", false, "Fail if .env contains keys not defined in schema")
	cmd.Flags().StringVar(&opts.envName, "env-name", "", "Environment name (e.g. production, development) for environment-specific rules")
	cmd.Flags().BoolVar(&opts.scanSecrets, "scan-secrets", false, "Scan for hardcoded secrets in .env values")
	cmd.Flags().StringVar(&opts.configPath, "config", "", "Path to config file (default: auto-discover)")
	cmd.Flags().BoolVar(&opts.failOnWarnings, "fail-on-warnings", false, "Treat warnings as errors")

	return cmd
}

func loadValidateConfig(opts *validateOptions) (*config.Config, error) {
	// Start with defaults
	cfg := config.Default()

	// Load config file if available
	var cfgPath string
	var found bool
	if opts.configPath != "" {
		cfgPath = opts.configPath
		found = true
	} else {
		wd, _ := os.Getwd()
		cfgPath, found = config.Find(wd)
	}

	if found {
		loaded, err := config.Load(cfgPath)
		if err != nil {
			return nil, fmt.Errorf("config error: %w", err)
		}
		cfg = config.Merge(cfg, loaded)
	}

	// Apply environment variable overrides
	cfg = config.EnvOverride(cfg)

	// Apply CLI flags (highest precedence)
	if opts.schemaPath != "" {
		cfg.Schema = opts.schemaPath
	}
	if len(opts.envPaths) > 0 {
		cfg.Env = opts.envPaths
	}
	if opts.format != "" {
		cfg.Format = opts.format
	}
	if opts.strict {
		cfg.Strict = true
	}
	if opts.envName != "" {
		cfg.EnvName = opts.envName
	}
	if opts.scanSecrets {
		cfg.ScanSecrets = true
	}
	if opts.failOnWarnings {
		cfg.FailOnWarnings = true
	}

	return cfg, nil
}

func runValidate(stdout, stderr io.Writer, opts *validateOptions) error {
	cfg, err := loadValidateConfig(opts)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return ErrIO
	}

	// Parse schema
	s, err := schema.Parse(cfg.Schema)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return ErrIO
	}

	// Parse .env files (later files override earlier ones)
	envVars := make(map[string]string)
	for _, path := range cfg.Env {
		vars, err := dotenv.Parse(path)
		if err != nil {
			fmt.Fprintf(stderr, "Error: %v\n", err)
			return ErrIO
		}
		for k, v := range vars {
			envVars[k] = v
		}
	}

	// Expand variable references
	if err := dotenv.Expand(envVars); err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return ErrValidationFailed
	}

	// Validate
	result := validator.Validate(s, envVars, cfg.Strict, cfg.EnvName)

	// Redact sensitive values from output
	result.RedactSensitive(envVars, s)

	// Scan for secrets if requested
	if cfg.ScanSecrets {
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
			sev := validator.SeverityError
			switch m.Severity {
			case secrets.SeverityCritical, secrets.SeverityHigh:
				sev = validator.SeverityError
			case secrets.SeverityMedium:
				sev = validator.SeverityWarn
			case secrets.SeverityLow:
				sev = validator.SeverityInfo
			}
			result.AddErrorWithSeverity(m.Key, "secret", m.Message+" (redacted: "+m.Redacted+")", sev)
		}
	}

	// Report
	switch cfg.Format {
	case "json":
		if err := reporter.JSON(stdout, result); err != nil {
			fmt.Fprintf(stderr, "Error: failed to format output: %v\n", err)
			return ErrIO
		}
	case "sarif":
		if err := reporter.SARIF(stdout, result, cfg.Env, version); err != nil {
			fmt.Fprintf(stderr, "Error: failed to format output: %v\n", err)
			return ErrIO
		}
	case "github":
		reporter.GitHub(stdout, result, cfg.Env)
	case "text":
		reporter.Text(stdout, result)
	default:
		fmt.Fprintf(stderr, "Error: unknown format %q\n", cfg.Format)
		return ErrIO
	}

	if !result.IsValid(cfg.FailOnWarnings) {
		return ErrValidationFailed
	}

	return nil
}
