package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"coingecko-cli/internal/display"
)

func printResult(jsonMode bool, headers []string, rows [][]string) {
	if jsonMode {
		printJSON(headers, rows)
		return
	}
	display.PrintTable(headers, rows)
}

func printJSON(headers []string, rows [][]string) {
	result := make([]map[string]string, len(rows))
	for i, row := range rows {
		obj := make(map[string]string, len(headers))
		for j, h := range headers {
			if j < len(row) {
				obj[h] = row[j]
			}
		}
		result[i] = obj
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(result)
}

func printJSONRaw(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func warnf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
}
