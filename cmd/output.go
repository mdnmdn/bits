package cmd

import (
	"encoding/json"
	"fmt"
	"os"
)

func printJSONRaw(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func warnf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
}
