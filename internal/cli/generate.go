package cli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/envguard/envguard/internal/schema"
)

func newGenerateExampleCmd() *cobra.Command {
	var schemaPath string
	var outputPath string

	cmd := &cobra.Command{
		Use:   "generate-example",
		Short: "Generate a .env.example file from the schema",
		Long:  `Reads the schema and outputs a .env.example file with placeholders and comments.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerateExample(cmd.OutOrStdout(), cmd.ErrOrStderr(), schemaPath, outputPath)
		},
		SilenceUsage: true,
	}

	cmd.Flags().StringVarP(&schemaPath, "schema", "s", "envguard.yaml", "Path to schema YAML file")
	cmd.Flags().StringVarP(&outputPath, "output", "o", ".env.example", "Path to output .env.example file")

	return cmd
}

func runGenerateExample(stdout, stderr io.Writer, schemaPath string, outputPath string) error {
	s, err := schema.Parse(schemaPath)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return ErrIO
	}

	if _, err := os.Stat(outputPath); err == nil {
		return fmt.Errorf("%s already exists", outputPath)
	}

	var lines []string

	// Sort keys for deterministic output
	keys := make([]string, 0, len(s.Env))
	for name := range s.Env {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	for _, name := range keys {
		v := s.Env[name]

		if v.Description != "" {
			lines = append(lines, fmt.Sprintf("# %s", v.Description))
		}

		if v.DevOnly {
			lines = append(lines, "# (development only)")
		}

		if len(v.RequiredIn) > 0 {
			lines = append(lines, fmt.Sprintf("# required in: %s", strings.Join(v.RequiredIn, ", ")))
		}

		placeholder := generatePlaceholder(name, v)
		lines = append(lines, fmt.Sprintf("%s=%s", name, placeholder))
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		fmt.Fprintf(stderr, "Error: failed to write %s: %v\n", outputPath, err)
		return ErrIO
	}

	fmt.Fprintf(stdout, "Created %s with %d variables\n", outputPath, len(keys))
	return nil
}

func generatePlaceholder(name string, v *schema.Variable) string {
	if v.Default != nil {
		return fmt.Sprintf("%v", v.Default)
	}

	if v.DevOnly {
		return "your-dev-value"
	}

	if len(v.Enum) > 0 {
		var vals []string
		for _, ev := range v.Enum {
			vals = append(vals, fmt.Sprintf("%v", ev))
		}
		return vals[0]
	}

	switch v.Type {
	case schema.TypeString:
		if v.Format == "email" {
			return "user@example.com"
		}
		if v.Format == "url" {
			return "https://example.com"
		}
		if v.Format == "uuid" {
			return "00000000-0000-0000-0000-000000000000"
		}
		if v.Pattern != "" {
			return "your-value"
		}
		return "your-value"
	case schema.TypeInteger:
		if v.Min != nil {
			return fmt.Sprintf("%v", v.Min)
		}
		return "0"
	case schema.TypeFloat:
		if v.Min != nil {
			return fmt.Sprintf("%v", v.Min)
		}
		return "0.0"
	case schema.TypeBoolean:
		return "false"
	default:
		return "your-value"
	}
}
