package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/envguard/envguard/internal/audit"
	"github.com/envguard/envguard/internal/reporter"
)

type auditOptions struct {
	srcDir     string
	envPath    string
	schemaPath string
	format     string
	exclude    []string
	ignoreVars []string
	strict     bool
}

func newAuditCmd() *cobra.Command {
	opts := &auditOptions{}

	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Audit source code for environment variable usage",
		Long: `Audit scans source code for environment variable references and compares
them against the .env file and schema. It reports missing, unused, and undocumented variables.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAudit(cmd.OutOrStdout(), cmd.ErrOrStderr(), opts)
		},
		SilenceUsage: true,
	}

	cmd.Flags().StringVar(&opts.srcDir, "src", ".", "Source directory to scan")
	cmd.Flags().StringVarP(&opts.envPath, "env", "e", ".env", "Path to .env file")
	cmd.Flags().StringVarP(&opts.schemaPath, "schema", "s", "", "Path to schema YAML file")
	cmd.Flags().StringVarP(&opts.format, "format", "f", "text", "Output format: text, json, or sarif")
	cmd.Flags().StringArrayVar(&opts.exclude, "exclude", []string{"vendor/", "node_modules/", ".git/", "dist/", "build/", "target/", ".venv/", "__pycache__/"}, "Patterns to exclude from scanning")
	cmd.Flags().StringArrayVar(&opts.ignoreVars, "ignore-var", nil, "Variables to ignore (can be specified multiple times)")
	cmd.Flags().BoolVar(&opts.strict, "strict", false, "Fail if missing or undocumented variables are found")

	return cmd
}

func runAudit(stdout, stderr io.Writer, opts *auditOptions) error {
	ignoreVars := opts.ignoreVars
	if len(ignoreVars) == 0 {
		ignoreVars = audit.KnownRuntimeVars()
	}

	result, err := audit.Run(audit.Options{
		SrcDir:     opts.srcDir,
		EnvPath:    opts.envPath,
		SchemaPath: opts.schemaPath,
		Exclude:    opts.exclude,
		IgnoreVars: ignoreVars,
		Strict:     opts.strict,
	})
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return ErrIO
	}

	if !result.HasFindings() {
		fmt.Fprintln(stdout, "✓ Audit complete. No issues found.")
		return nil
	}

	switch opts.format {
	case "json":
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(result.Findings); err != nil {
			fmt.Fprintf(stderr, "Error: failed to format output: %v\n", err)
			return ErrIO
		}
	case "sarif":
		if err := reporter.SARIFAudit(stdout, result, version); err != nil {
			fmt.Fprintf(stderr, "Error: failed to format output: %v\n", err)
			return ErrIO
		}
	case "text":
		printAuditText(stdout, result)
	default:
		fmt.Fprintf(stderr, "Error: unknown format %q\n", opts.format)
		return ErrIO
	}

	if opts.strict {
		for _, f := range result.Findings {
			if f.Type == audit.MissingVar || f.Type == audit.UndocumentedVar || f.Type == audit.MissingRequired {
				return ErrValidationFailed
			}
		}
	}

	return nil
}

func printAuditText(w io.Writer, result *audit.Result) {
	// Group findings by type
	byType := make(map[audit.FindingType][]audit.Finding)
	for _, f := range result.Findings {
		byType[f.Type] = append(byType[f.Type], f)
	}

	typeNames := map[audit.FindingType]string{
		audit.MissingVar:      "Missing",
		audit.UnusedVar:       "Unused",
		audit.UndocumentedVar: "Undocumented",
		audit.MissingRequired: "Missing Required",
	}

	typeOrder := []audit.FindingType{audit.MissingRequired, audit.MissingVar, audit.UndocumentedVar, audit.UnusedVar}
	typeSymbols := map[audit.FindingType]string{
		audit.MissingVar:      "✗",
		audit.UnusedVar:       "⚠",
		audit.UndocumentedVar: "⚠",
		audit.MissingRequired: "✗",
	}

	for _, ft := range typeOrder {
		findings, ok := byType[ft]
		if !ok {
			continue
		}
		fmt.Fprintf(w, "%s %s (%d):\n\n", typeSymbols[ft], typeNames[ft], len(findings))
		for _, f := range findings {
			if f.File != "" {
				fmt.Fprintf(w, "  • %s\n", f.Var)
				fmt.Fprintf(w, "    └─ %s\n", f.Message)
				if f.Line > 0 {
					fmt.Fprintf(w, "       at %s:%d\n", f.File, f.Line)
				}
			} else {
				fmt.Fprintf(w, "  • %s\n", f.Var)
				fmt.Fprintf(w, "    └─ %s\n", f.Message)
			}
		}
		fmt.Fprintln(w)
	}

	// Summary
	fmt.Fprintln(w, strings.Repeat("─", 40))
	total := len(result.Findings)
	fmt.Fprintf(w, "Total: %d finding(s)\n", total)
	for _, ft := range typeOrder {
		if c := result.CountByType(ft); c > 0 {
			fmt.Fprintf(w, "  • %s: %d\n", typeNames[ft], c)
		}
	}
}
