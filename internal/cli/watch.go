package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/envguard/envguard/internal/dotenv"
	"github.com/envguard/envguard/internal/reporter"
	"github.com/envguard/envguard/internal/schema"
	"github.com/envguard/envguard/internal/validator"
	"github.com/envguard/envguard/internal/watch"
)

type watchOptions struct {
	schemaPath string
	envPaths   []string
	format     string
	strict     bool
	envName    string
	debounce   time.Duration
	cmdSuccess string
	cmdFail    string
	quiet      bool
}

func newWatchCmd() *cobra.Command {
	opts := &watchOptions{}

	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Watch .env and schema files for changes and re-validate",
		Long: `Watch monitors .env and envguard.yaml files for changes.
When a change is detected, it re-runs validation automatically.
Press Ctrl+C to exit.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runWatch(cmd.OutOrStdout(), cmd.ErrOrStderr(), opts)
		},
		SilenceUsage: true,
	}

	cmd.Flags().StringVarP(&opts.schemaPath, "schema", "s", "envguard.yaml", "Path to schema YAML file")
	cmd.Flags().StringArrayVarP(&opts.envPaths, "env", "e", []string{".env"}, "Path to .env file (can be specified multiple times)")
	cmd.Flags().StringVarP(&opts.format, "format", "f", "text", "Output format: text, json, github, or sarif")
	cmd.Flags().BoolVar(&opts.strict, "strict", false, "Fail if .env contains keys not defined in schema")
	cmd.Flags().StringVar(&opts.envName, "env-name", "", "Environment name for environment-specific rules")
	cmd.Flags().DurationVar(&opts.debounce, "debounce", 300*time.Millisecond, "Debounce duration for file changes")
	cmd.Flags().StringVar(&opts.cmdSuccess, "cmd", "", "Command to run on validation success")
	cmd.Flags().StringVar(&opts.cmdFail, "cmd-on-fail", "", "Command to run on validation failure")
	cmd.Flags().BoolVar(&opts.quiet, "quiet", false, "Only show errors")

	return cmd
}

func runWatch(stdout, stderr io.Writer, opts *watchOptions) error {
	// Collect all paths to watch
	watchPaths := []string{opts.schemaPath}
	watchPaths = append(watchPaths, opts.envPaths...)

	// Resolve paths
	resolvedPaths := make([]string, 0, len(watchPaths))
	for _, p := range watchPaths {
		abs, err := os.Getwd()
		if err != nil {
			abs = "."
		}
		if !filepath.IsAbs(p) {
			p = filepath.Join(abs, p)
		}
		resolvedPaths = append(resolvedPaths, p)
	}

	w := watch.New(watch.Options{
		Paths:      resolvedPaths,
		Debounce:   opts.debounce,
		CmdSuccess: opts.cmdSuccess,
		CmdFail:    opts.cmdFail,
		Quiet:      opts.quiet,
	})

	var validationCount atomic.Int32
	var lastValid atomic.Bool

	w.SetCallback(func() error {
		validationCount.Add(1)

		// Parse schema
		s, err := schema.Parse(opts.schemaPath)
		if err != nil {
			if !opts.quiet {
				fmt.Fprintf(stderr, "Error: failed to parse schema: %v\n", err)
			}
			lastValid.Store(false)
			return err
		}

		// Parse .env files
		envVars := make(map[string]string)
		for _, path := range opts.envPaths {
			vars, err := dotenv.Parse(path)
			if err != nil {
				if !opts.quiet {
					fmt.Fprintf(stderr, "Error: failed to parse %s: %v\n", path, err)
				}
				lastValid.Store(false)
				return err
			}
			for k, v := range vars {
				envVars[k] = v
			}
		}

		// Expand variables
		if err := dotenv.Expand(envVars); err != nil {
			if !opts.quiet {
				fmt.Fprintf(stderr, "Error: variable expansion failed: %v\n", err)
			}
			lastValid.Store(false)
			return err
		}

		// Validate
		result := validator.Validate(s, envVars, opts.strict, opts.envName)
		result.RedactSensitive(envVars, s)

		// Report
		switch opts.format {
		case "json":
			reporter.JSON(stdout, result)
		case "sarif":
			reporter.SARIF(stdout, result, opts.envPaths, version)
		case "github":
			reporter.GitHub(stdout, result, opts.envPaths)
		default:
			reporter.Text(stdout, result)
		}

		if !opts.quiet {
			fmt.Fprintln(stdout)
			fmt.Fprintf(stdout, "─ Run #%d ─ %s\n", validationCount.Load(), time.Now().Format("15:04:05"))
			if result.Valid {
				fmt.Fprintln(stdout, "Status: ✓ Valid")
			} else {
				fmt.Fprintf(stdout, "Status: ✗ Invalid (%d error(s), %d warning(s))\n", len(result.Errors), len(result.Warnings))
			}
			fmt.Fprintln(stdout, "Press Ctrl+C to exit")
		}

		lastValid.Store(result.Valid)
		if !result.Valid {
			return fmt.Errorf("validation failed")
		}
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		if !opts.quiet {
			fmt.Fprintln(stdout, "\n\nShutting down watch mode...")
			fmt.Fprintf(stdout, "Total validations: %d\n", validationCount.Load())
			if validationCount.Load() > 0 {
				if lastValid.Load() {
					fmt.Fprintln(stdout, "Last run: ✓ Valid")
				} else {
					fmt.Fprintln(stdout, "Last run: ✗ Invalid")
				}
			}
		}
		cancel()
	}()

	if !opts.quiet {
		fmt.Fprintln(stdout, "Starting EnvGuard watch mode...")
		fmt.Fprintf(stdout, "Watching: %v\n", watchPaths)
		fmt.Fprintf(stdout, "Debounce: %v\n", opts.debounce)
		fmt.Fprintln(stdout)
	}

	return w.Run(ctx)
}
