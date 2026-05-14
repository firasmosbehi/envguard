package reporter

import (
	"fmt"
	"io"

	"github.com/envguard/envguard/internal/validator"
)

func severitySymbol(sev validator.Severity) string {
	switch sev {
	case validator.SeverityWarn:
		return "⚠"
	case validator.SeverityInfo:
		return "ℹ"
	default:
		return "✗"
	}
}

// Text writes a human-readable validation report to w.
func Text(w io.Writer, result *validator.Result) {
	hasErrors := false
	hasWarnings := false
	hasInfos := false

	for _, err := range result.Errors {
		switch err.Severity {
		case validator.SeverityWarn:
			hasWarnings = true
		case validator.SeverityInfo:
			hasInfos = true
		default:
			hasErrors = true
		}
	}

	if !result.Valid {
		hasErrors = true
	}

	if !hasErrors && len(result.Warnings) == 0 && !hasWarnings && !hasInfos {
		fmt.Fprintln(w, "✓ All environment variables validated.")
		return
	}

	if hasErrors {
		errorCount := 0
		for _, err := range result.Errors {
			if err.Severity == validator.SeverityError {
				errorCount++
			}
		}
		fmt.Fprintf(w, "✗ Environment validation failed (%d error(s))\n\n", errorCount)
		for _, err := range result.Errors {
			if err.Severity == validator.SeverityError {
				fmt.Fprintf(w, "  • %s\n", err.Key)
				fmt.Fprintf(w, "    └─ %s: %s\n", err.Rule, err.Message)
			}
		}
	}

	if hasWarnings {
		if hasErrors {
			fmt.Fprintln(w)
		}
		warningCount := 0
		for _, err := range result.Errors {
			if err.Severity == validator.SeverityWarn {
				warningCount++
			}
		}
		fmt.Fprintf(w, "⚠ Warnings (%d):\n\n", warningCount)
		for _, err := range result.Errors {
			if err.Severity == validator.SeverityWarn {
				fmt.Fprintf(w, "  • %s\n", err.Key)
				fmt.Fprintf(w, "    └─ %s: %s\n", err.Rule, err.Message)
			}
		}
	}

	if hasInfos {
		if hasErrors || hasWarnings {
			fmt.Fprintln(w)
		}
		infoCount := 0
		for _, err := range result.Errors {
			if err.Severity == validator.SeverityInfo {
				infoCount++
			}
		}
		fmt.Fprintf(w, "ℹ Info (%d):\n\n", infoCount)
		for _, err := range result.Errors {
			if err.Severity == validator.SeverityInfo {
				fmt.Fprintf(w, "  • %s\n", err.Key)
				fmt.Fprintf(w, "    └─ %s: %s\n", err.Rule, err.Message)
			}
		}
	}

	if len(result.Warnings) > 0 {
		if hasErrors || hasWarnings || hasInfos {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "⚠ General Warnings (%d):\n\n", len(result.Warnings))
		for _, warn := range result.Warnings {
			fmt.Fprintf(w, "  • %s\n", warn.Key)
			fmt.Fprintf(w, "    └─ %s: %s\n", warn.Rule, warn.Message)
		}
	}
}
