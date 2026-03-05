package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"coingecko-cli/internal/export"
)

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
