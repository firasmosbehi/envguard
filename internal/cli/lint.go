package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/envguard/envguard/internal/reporter"
	"github.com/envguard/envguard/internal/schema"
)

type lintOptions struct {
	schemaPath string
	format     string
}

func newLintCmd() *cobra.Command {
	opts := &lintOptions{}

	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Lint an EnvGuard schema file for best practices",
		Long:  `Lint checks a schema YAML file for structural issues, redundancies, and best practice violations.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runLint(cmd.OutOrStdout(), cmd.ErrOrStderr(), opts)
		},
		SilenceUsage: true,
	}

	cmd.Flags().StringVarP(&opts.schemaPath, "schema", "s", "envguard.yaml", "Path to schema YAML file")
	cmd.Flags().StringVarP(&opts.format, "format", "f", "text", "Output format: text, json, or sarif")

	return cmd
}

func runLint(stdout, stderr io.Writer, opts *lintOptions) error {
	s, err := schema.ParseLenient(opts.schemaPath)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return ErrIO
	}

	findings := lintSchema(s)

	if len(findings) == 0 {
		fmt.Fprintln(stdout, "✓ Schema passes all lint checks.")
		return nil
	}

	switch opts.format {
	case "json":
		fmt.Fprintln(stdout, "[")
		for i, f := range findings {
			comma := ","
			if i == len(findings)-1 {
				comma = ""
			}
			fmt.Fprintf(stdout, `  {"level": %q, "rule": %q, "message": %q}%s`+"\n", f.Level, f.Rule, f.Message, comma)
		}
		fmt.Fprintln(stdout, "]")
	case "sarif":
		if err := reporter.SARIFLint(stdout, findings, version); err != nil {
			fmt.Fprintf(stderr, "Error: failed to format output: %v\n", err)
			return ErrIO
		}
	case "text":
		fmt.Fprintf(stdout, "✗ Schema lint found %d issue(s)\n\n", len(findings))
		for _, f := range findings {
			symbol := "•"
			switch f.Level {
			case "error":
				symbol = "✗"
			case "warning":
				symbol = "⚠"
			}
			fmt.Fprintf(stdout, "  %s %s\n", symbol, f.Message)
		}
	default:
		fmt.Fprintf(stderr, "Error: unknown format %q\n", opts.format)
		return ErrIO
	}

	return ErrValidationFailed
}

func lintSchema(s *schema.Schema) []reporter.LintFinding {
	var findings []reporter.LintFinding

	for name, v := range s.Env {
		// Redundant rules
		if v.Required && v.Default != nil {
			findings = append(findings, reporter.LintFinding{
				Level:   "error",
				Rule:    "redundant",
				Message: fmt.Sprintf("variable %q: required and default are mutually exclusive", name),
			})
		}

		if v.AllowEmpty != nil && !*v.AllowEmpty && v.Required {
			findings = append(findings, reporter.LintFinding{
				Level:   "warning",
				Rule:    "redundant",
				Message: fmt.Sprintf("variable %q: allowEmpty=false is redundant when required=true", name),
			})
		}

		// Check for unreachable dependsOn
		if v.DependsOn != "" {
			if _, exists := s.Env[v.DependsOn]; !exists {
				findings = append(findings, reporter.LintFinding{
					Level:   "error",
					Rule:    "unreachable",
					Message: fmt.Sprintf("variable %q: dependsOn references undefined variable %q", name, v.DependsOn),
				})
			}
		}

		// Check for empty enum
		if v.Enum != nil && len(v.Enum) == 0 {
			findings = append(findings, reporter.LintFinding{
				Level:   "error",
				Rule:    "empty-enum",
				Message: fmt.Sprintf("variable %q: enum is empty (no values allowed)", name),
			})
		}

		// Check for pattern on non-string type (schema.Validate should catch this, but lint is explicit)
		if v.Pattern != "" && v.Type != schema.TypeString {
			findings = append(findings, reporter.LintFinding{
				Level:   "error",
				Rule:    "type-mismatch",
				Message: fmt.Sprintf("variable %q: pattern can only be used with string type", name),
			})
		}

		// Warn about missing description
		if v.Description == "" {
			findings = append(findings, reporter.LintFinding{
				Level:   "warning",
				Rule:    "missing-description",
				Message: fmt.Sprintf("variable %q: missing description", name),
			})
		}

		// Check for suspicious defaults
		if v.Default != nil {
			defStr := fmt.Sprintf("%v", v.Default)
			if strings.Contains(strings.ToLower(defStr), "changeme") ||
				strings.Contains(strings.ToLower(defStr), "password") ||
				strings.Contains(strings.ToLower(defStr), "secret") ||
				strings.Contains(strings.ToLower(defStr), "placeholder") {
				findings = append(findings, reporter.LintFinding{
					Level:   "warning",
					Rule:    "suspicious-default",
					Message: fmt.Sprintf("variable %q: default value %q looks suspicious", name, defStr),
				})
			}
		}

		// Check min > max
		if v.Min != nil && v.Max != nil {
			minVal, _ := toFloat64(v.Min)
			maxVal, _ := toFloat64(v.Max)
			if minVal > maxVal {
				findings = append(findings, reporter.LintFinding{
					Level:   "error",
					Rule:    "range",
					Message: fmt.Sprintf("variable %q: min (%v) is greater than max (%v)", name, v.Min, v.Max),
				})
			}
		}

		// Check minLength > maxLength
		if v.MinLength != nil && v.MaxLength != nil && *v.MinLength > *v.MaxLength {
			findings = append(findings, reporter.LintFinding{
				Level:   "error",
				Rule:    "range",
				Message: fmt.Sprintf("variable %q: minLength (%d) is greater than maxLength (%d)", name, *v.MinLength, *v.MaxLength),
			})
		}

		// Warn if deprecated but no replacement suggested
		if v.Deprecated != "" && !strings.Contains(strings.ToLower(v.Deprecated), "use") {
			findings = append(findings, reporter.LintFinding{
				Level:   "warning",
				Rule:    "deprecated",
				Message: fmt.Sprintf("variable %q: deprecated without suggesting a replacement", name),
			})
		}
	}

	return findings
}

func toFloat64(value any) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}
