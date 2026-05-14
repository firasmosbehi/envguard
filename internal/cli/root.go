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

// version is set at runtime by Execute and used by reporters (e.g., SARIF).
var version string

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute(v string) error {
	version = v
	rootCmd.AddCommand(newValidateCmd())
	rootCmd.AddCommand(newScanCmd())
	rootCmd.AddCommand(newLintCmd())
	rootCmd.AddCommand(newAuditCmd())
	rootCmd.AddCommand(newSyncCmd())
	rootCmd.AddCommand(newWatchCmd())
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newGenerateExampleCmd())
	rootCmd.AddCommand(newInstallHookCmd())
	rootCmd.AddCommand(newUninstallHookCmd())
	rootCmd.AddCommand(newDocsCmd())
	rootCmd.AddCommand(newLSPCmd())
	rootCmd.AddCommand(newVersionCmd(v))
	return rootCmd.Execute()
}
