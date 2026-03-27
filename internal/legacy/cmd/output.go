package legacycmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/mdnmdn/bits/internal/legacy/model"
	"github.com/mdnmdn/bits/internal/legacy/export"

	"github.com/spf13/cobra"
)

// CLIError is the structured JSON error format emitted to stderr when -o json is set.
type CLIError struct {
	Error      string `json:"error"`
	Message    string `json:"message"`
	RetryAfter *int   `json:"retry_after,omitempty"`
}

func printJSONRaw(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func warnf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
}

func exportCSV(path string, headers []string, rows [][]string) error {
	if err := export.ExportCSV(path, headers, rows); err != nil {
		return err
	}
	warnf("Exported to %s\n", path)
	return nil
}

// formatError writes a structured JSON error to stderr when -o json is active,
// otherwise returns the error unchanged for Cobra's default plain text handling.
func formatError(cmd *cobra.Command, err error) error {
	if !outputJSON(cmd) {
		return err
	}

	cliErr := classifyError(err)
	enc := json.NewEncoder(os.Stderr)
	enc.SetIndent("", "  ")
	_ = enc.Encode(cliErr)
	return err
}

func classifyError(err error) CLIError {
	var rle *model.RateLimitError
	if errors.As(err, &rle) {
		ce := CLIError{Error: "rate_limited", Message: rle.Error()}
		if rle.RetryAfter > 0 {
			ce.RetryAfter = &rle.RetryAfter
		}
		return ce
	}
	if errors.Is(err, model.ErrInvalidAPIKey) {
		return CLIError{Error: "invalid_api_key", Message: err.Error()}
	}
	if errors.Is(err, model.ErrPlanRestricted) {
		return CLIError{Error: "plan_restricted", Message: err.Error()}
	}
	return CLIError{Error: "error", Message: err.Error()}
}
