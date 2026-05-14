package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/envguard/envguard/internal/docs"
	"github.com/envguard/envguard/internal/schema"
)

type docsOptions struct {
	schemaPath string
	format     string
	output     string
	groupBy    string
}

func newDocsCmd() *cobra.Command {
	opts := &docsOptions{}

	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Generate documentation from schema",
		Long:  `Generate Markdown, HTML, or JSON documentation from an EnvGuard schema file.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runDocs(cmd.OutOrStdout(), cmd.ErrOrStderr(), opts)
		},
		SilenceUsage: true,
	}

	cmd.Flags().StringVarP(&opts.schemaPath, "schema", "s", "envguard.yaml", "Path to schema YAML file")
	cmd.Flags().StringVarP(&opts.format, "format", "f", "markdown", "Output format: markdown, html, or json")
	cmd.Flags().StringVarP(&opts.output, "output", "o", "", "Output file path (default: stdout)")
	cmd.Flags().StringVar(&opts.groupBy, "group-by", "", "Group variables by: prefix")

	return cmd
}

func runDocs(stdout, stderr io.Writer, opts *docsOptions) error {
	s, err := schema.Parse(opts.schemaPath)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return ErrIO
	}

	content, err := docs.Generate(s, docs.Options{
		Format:  opts.format,
		GroupBy: opts.groupBy,
	})
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return ErrIO
	}

	if opts.output != "" {
		if err := os.WriteFile(opts.output, []byte(content), 0644); err != nil {
			fmt.Fprintf(stderr, "Error: failed to write output: %v\n", err)
			return ErrIO
		}
		fmt.Fprintf(stdout, "✓ Generated %s documentation at %s\n", opts.format, opts.output)
	} else {
		fmt.Fprint(stdout, content)
	}

	return nil
}
