package reporter

import (
	"fmt"
	"io"

	"github.com/envguard/envguard/internal/validator"
)

// GitHub writes validation results as GitHub Actions workflow commands.
// See: https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions
func GitHub(w io.Writer, result *validator.Result, envPaths []string) {
	if result.Valid && len(result.Warnings) == 0 {
		fmt.Fprintln(w, "✓ All environment variables validated.")
		return
	}

	for _, err := range result.Errors {
		msg := fmt.Sprintf("%s: %s", err.Rule, err.Message)
		fmt.Fprintf(w, "::error title=EnvGuard Validation Error::%s=%s\n", err.Key, msg)
	}

	for _, warn := range result.Warnings {
		msg := fmt.Sprintf("%s: %s", warn.Rule, warn.Message)
		fmt.Fprintf(w, "::warning title=EnvGuard Validation Warning::%s=%s\n", warn.Key, msg)
	}

	if !result.Valid {
		fmt.Fprintf(w, "\n✗ Environment validation failed (%d error(s), %d warning(s))\n", len(result.Errors), len(result.Warnings))
	} else {
		fmt.Fprintf(w, "\n⚠ Environment validation passed with %d warning(s)\n", len(result.Warnings))
	}
}
