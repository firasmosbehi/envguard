package reporter

import (
	"encoding/json"
	"io"

	"github.com/envguard/envguard/internal/validator"
)

// JSON writes a machine-readable validation report to w.
func JSON(w io.Writer, result *validator.Result) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
