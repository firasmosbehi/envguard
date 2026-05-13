package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const sampleSchema = `version: "1.0"

env:
  DATABASE_URL:
    type: string
    required: true
    description: "Database connection string"

  PORT:
    type: integer
    min: 1024
    max: 65535
    default: 3000
    description: "Server port"

  DEBUG:
    type: boolean
    default: false
    description: "Enable debug mode"

  LOG_LEVEL:
    type: string
    enum: [debug, info, warn, error]
    default: info
    description: "Logging verbosity"

  API_KEY:
    type: string
    required: true
    minLength: 32
    pattern: "^[A-Za-z0-9_-]+$"
    description: "API authentication key"
`

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Generate a sample envguard.yaml schema file",
		Long:  `Creates a starter envguard.yaml file in the current directory.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path := "envguard.yaml"
			if _, err := os.Stat(path); err == nil {
				return fmt.Errorf("%s already exists", path)
			}

			if err := os.WriteFile(path, []byte(sampleSchema), 0644); err != nil {
				return fmt.Errorf("failed to write %s: %w", path, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", path)
			return nil
		},
	}
}
