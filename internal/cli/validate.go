package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/envguard/envguard/internal/dotenv"
	"github.com/envguard/envguard/internal/reporter"
	"github.com/envguard/envguard/internal/schema"
	"github.com/envguard/envguard/internal/validator"
)

type validateOptions struct {
	schemaPath string
	envPath    string
	format     string
	strict     bool
}

func newValidateCmd() *cobra.Command {
	opts := &validateOptions{}

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a .env file against a schema",
		Long:  `Validate checks that the given .env file satisfies the rules defined in the schema YAML file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(cmd.OutOrStdout(), cmd.ErrOrStderr(), opts)
		},
		SilenceUsage: true,
	}

	cmd.Flags().StringVarP(&opts.schemaPath, "schema", "s", "envguard.yaml", "Path to schema YAML file")
	cmd.Flags().StringVarP(&opts.envPath, "env", "e", ".env", "Path to .env file")
	cmd.Flags().StringVarP(&opts.format, "format", "f", "text", "Output format: text or json")
	cmd.Flags().BoolVar(&opts.strict, "strict", false, "Fail if .env contains keys not defined in schema")

	return cmd
}

func runValidate(stdout, stderr io.Writer, opts *validateOptions) error {
	// Parse schema
	s, err := schema.Parse(opts.schemaPath)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return ErrIO
	}

	// Parse .env file
	envVars, err := dotenv.Parse(opts.envPath)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return ErrIO
	}

	// Validate
	result := validator.Validate(s, envVars, opts.strict)

	// Report
	switch opts.format {
	case "json":
		if err := reporter.JSON(stdout, result); err != nil {
			fmt.Fprintf(stderr, "Error: failed to format output: %v\n", err)
			return ErrIO
		}
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
