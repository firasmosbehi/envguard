package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/envguard/envguard/internal/lsp"
)

func newLSPCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lsp",
		Short: "Start the EnvGuard Language Server",
		Long:  `Starts an LSP server for real-time .env validation in editors.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.ErrOrStderr(), "EnvGuard LSP server starting...")
			server := lsp.NewServer(os.Stdin, os.Stdout)
			return server.Run()
		},
		SilenceUsage: true,
	}
}
