package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/envguard/envguard/internal/hooks"
)

func newInstallHookCmd() *cobra.Command {
	opts := &struct {
		hookType string
		force    bool
		command  string
	}{}

	cmd := &cobra.Command{
		Use:   "install-hook",
		Short: "Install a Git hook for EnvGuard",
		Long:  `Installs a pre-commit or pre-push hook that runs EnvGuard validation automatically.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runInstallHook(cmd.OutOrStdout(), cmd.ErrOrStderr(), opts)
		},
		SilenceUsage: true,
	}

	cmd.Flags().StringVar(&opts.hookType, "type", "pre-commit", "Hook type: pre-commit or pre-push")
	cmd.Flags().BoolVar(&opts.force, "force", false, "Overwrite existing hook")
	cmd.Flags().StringVar(&opts.command, "command", "", "Custom command to run (default: envguard validate --strict)")

	return cmd
}

func newUninstallHookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall-hook",
		Short: "Uninstall a Git hook",
		RunE: func(cmd *cobra.Command, args []string) error {
			hookType := "pre-commit"
			if len(args) > 0 {
				hookType = args[0]
			}
			if err := hooks.Uninstall(hookType); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return ErrIO
			}
			fmt.Fprintf(cmd.OutOrStdout(), "✓ Uninstalled %s hook\n", hookType)
			return nil
		},
		SilenceUsage: true,
	}
	return cmd
}

func runInstallHook(stdout, stderr io.Writer, opts *struct {
	hookType string
	force    bool
	command  string
}) error {
	if err := hooks.Install(hooks.Options{
		HookType: opts.hookType,
		Force:    opts.force,
		Command:  opts.command,
	}); err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return ErrIO
	}
	fmt.Fprintf(stdout, "✓ Installed %s hook\n", opts.hookType)
	return nil
}
