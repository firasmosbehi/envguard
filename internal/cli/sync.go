package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/envguard/envguard/internal/reporter"
	"github.com/envguard/envguard/internal/sync"
)

type syncOptions struct {
	envPath     string
	examplePath string
	schemaPath  string
	format      string
	check       bool
	addMissing  bool
}

func newSyncCmd() *cobra.Command {
	opts := &syncOptions{}

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync .env and .env.example files",
		Long: `Sync ensures that .env.example stays up-to-date with .env and the schema.
It can also check for drift (CI mode) and add missing keys to .env.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSync(cmd.OutOrStdout(), cmd.ErrOrStderr(), opts)
		},
		SilenceUsage: true,
	}

	cmd.Flags().StringVarP(&opts.envPath, "env", "e", ".env", "Path to .env file")
	cmd.Flags().StringVar(&opts.examplePath, "example", ".env.example", "Path to .env.example file")
	cmd.Flags().StringVarP(&opts.schemaPath, "schema", "s", "", "Path to schema YAML file")
	cmd.Flags().StringVarP(&opts.format, "format", "f", "text", "Output format: text, json, or sarif")
	cmd.Flags().BoolVar(&opts.check, "check", false, "Check for drift without writing (CI mode)")
	cmd.Flags().BoolVar(&opts.addMissing, "add-missing", false, "Add missing keys to .env with empty values")

	return cmd
}

func runSync(stdout, stderr io.Writer, opts *syncOptions) error {
	result, err := sync.Run(sync.Options{
		EnvPath:     opts.envPath,
		ExamplePath: opts.examplePath,
		SchemaPath:  opts.schemaPath,
		Check:       opts.check,
		AddMissing:  opts.addMissing,
	})
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return ErrIO
	}

	if len(result.Diffs) == 0 {
		fmt.Fprintln(stdout, "✓ .env and .env.example are in sync.")
		return nil
	}

	if opts.check {
		fmt.Fprintf(stderr, "✗ Drift detected between .env and .env.example\n\n")
	}

	switch opts.format {
	case "json":
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(result); err != nil {
			fmt.Fprintf(stderr, "Error: failed to format output: %v\n", err)
			return ErrIO
		}
	case "sarif":
		if err := reporter.SARIFSync(stdout, result, version); err != nil {
			fmt.Fprintf(stderr, "Error: failed to format output: %v\n", err)
			return ErrIO
		}
	case "text":
		printSyncText(stdout, result, opts.check)
	default:
		fmt.Fprintf(stderr, "Error: unknown format %q\n", opts.format)
		return ErrIO
	}

	if opts.check {
		return ErrValidationFailed
	}

	fmt.Fprintf(stdout, "✓ Updated %s (%d change(s))\n", opts.examplePath, len(result.Diffs))
	return nil
}

func printSyncText(w io.Writer, result *sync.Result, check bool) {
	if check {
		fmt.Fprintln(w, "The following changes would be made:")
		fmt.Fprintln(w)
	}

	for _, d := range result.Diffs {
		switch d.Type {
		case "missing-in-example":
			fmt.Fprintf(w, "  + %s=", d.Key)
			if d.EnvVal != "" {
				fmt.Fprintf(w, "%s", d.EnvVal)
			}
			fmt.Fprintln(w, "  (add to .env.example)")
		case "missing-in-env":
			fmt.Fprintf(w, "  - %s", d.Key)
			if d.ExVal != "" {
				fmt.Fprintf(w, "=%s", d.ExVal)
			}
			fmt.Fprintln(w, "  (present in .env.example but not .env)")
		}
	}
}
