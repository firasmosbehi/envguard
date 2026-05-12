package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "envguard",
	Short: "Validate .env files against a declarative YAML schema",
	Long: `EnvGuard is a CLI tool that validates .env files against a declarative YAML schema.
It catches missing, mistyped, or malformed environment variables before deployment.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute(version string) error {
	rootCmd.AddCommand(newValidateCmd())
	rootCmd.AddCommand(newScanCmd())
	rootCmd.AddCommand(newLintCmd())
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newGenerateExampleCmd())
	rootCmd.AddCommand(newVersionCmd(version))
	return rootCmd.Execute()
}
