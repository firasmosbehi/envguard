package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/envguard/envguard/internal/dotenv"
	"github.com/envguard/envguard/internal/infer"
)

func newInitCmd() *cobra.Command {
	opts := &struct {
		infer   bool
		envPath string
		config  bool
	}{}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generate a starter envguard.yaml schema file",
		Long:  `Creates a sample envguard.yaml in the current directory.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runInit(cmd.OutOrStdout(), cmd.ErrOrStderr(), opts)
		},
		SilenceUsage: true,
	}

	cmd.Flags().BoolVar(&opts.infer, "infer", false, "Infer schema from existing .env file")
	cmd.Flags().StringVarP(&opts.envPath, "env", "e", ".env", "Path to .env file for inference")
	cmd.Flags().BoolVar(&opts.config, "config", false, "Generate a .envguardrc.yaml config file")

	return cmd
}

func runInit(stdout, stderr io.Writer, opts *struct {
	infer   bool
	envPath string
	config  bool
}) error {
	if opts.config {
		return generateConfigFile(stdout, stderr)
	}

	if _, err := os.Stat("envguard.yaml"); err == nil {
		fmt.Fprintf(stderr, "Error: envguard.yaml already exists\n")
		return ErrValidationFailed
	}

	if opts.infer {
		envVars, err := dotenv.Parse(opts.envPath)
		if err != nil {
			fmt.Fprintf(stderr, "Error: failed to parse %s: %v\n", opts.envPath, err)
			return ErrIO
		}
		result := infer.FromEnv(envVars)
		yaml := result.GenerateYAML("1.0")
		if err := os.WriteFile("envguard.yaml", []byte(yaml), 0644); err != nil {
			fmt.Fprintf(stderr, "Error: failed to write envguard.yaml: %v\n", err)
			return ErrIO
		}
		fmt.Fprintf(stdout, "✓ Generated envguard.yaml from %s (%d variables inferred)\n", opts.envPath, len(result.Variables))
		return nil
	}

	// Default starter schema
	starter := `version: "1.0"
env:
  DATABASE_URL:
    type: string
    required: true
    description: "Database connection string"
    format: url

  PORT:
    type: integer
    default: 3000
    description: "Server port"
    min: 1
    max: 65535

  DEBUG:
    type: boolean
    default: false
    description: "Enable debug mode"
`
	if err := os.WriteFile("envguard.yaml", []byte(starter), 0644); err != nil {
		fmt.Fprintf(stderr, "Error: failed to write envguard.yaml: %v\n", err)
		return ErrIO
	}
	fmt.Fprintln(stdout, "✓ Created envguard.yaml")
	return nil
}

func generateConfigFile(stdout, stderr io.Writer) error {
	config := `schema: envguard.yaml
env:
  - .env
format: text
strict: false
`
	if err := os.WriteFile(".envguardrc.yaml", []byte(config), 0644); err != nil {
		fmt.Fprintf(stderr, "Error: failed to write .envguardrc.yaml: %v\n", err)
		return ErrIO
	}
	fmt.Fprintln(stdout, "✓ Created .envguardrc.yaml")
	return nil
}
