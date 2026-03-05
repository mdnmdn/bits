package export

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
)

func ExportCSV(path string, headers []string, rows [][]string) error {
	path = filepath.Clean(path)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating CSV file: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)

	if err := w.Write(headers); err != nil {
		return fmt.Errorf("writing CSV headers: %w", err)
	}
	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return fmt.Errorf("writing CSV row: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return fmt.Errorf("flushing CSV: %w", err)
	}
	return nil
}
