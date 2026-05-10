package reporter

import (
	"fmt"
	"io"

	"github.com/envguard/envguard/internal/validator"
)

// Text writes a human-readable validation report to w.
func Text(w io.Writer, result *validator.Result) {
	if result.Valid && len(result.Warnings) == 0 {
		fmt.Fprintln(w, "✓ All environment variables validated.")
		return
	}

	if !result.Valid {
		fmt.Fprintf(w, "✗ Environment validation failed (%d error(s))\n\n", len(result.Errors))
		for _, err := range result.Errors {
			fmt.Fprintf(w, "  • %s\n", err.Key)
			fmt.Fprintf(w, "    └─ %s: %s\n", err.Rule, err.Message)
		}
	}

	if len(result.Warnings) > 0 {
		if !result.Valid {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "⚠ Warnings (%d):\n\n", len(result.Warnings))
		for _, warn := range result.Warnings {
			fmt.Fprintf(w, "  • %s\n", warn.Key)
			fmt.Fprintf(w, "    └─ %s: %s\n", warn.Rule, warn.Message)
		}
	}
}
