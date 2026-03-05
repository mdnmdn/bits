package display

import (
	"fmt"
	"strings"
	"unicode"
)

func PrintTable(headers []string, rows [][]string) {
	if len(rows) == 0 {
		fmt.Println("No data to display.")
		return
	}

	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = VisibleWidth(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				w := VisibleWidth(cell)
				if w > widths[i] {
					widths[i] = w
				}
			}
		}
	}

	printRow(headers, widths)
	printSeparator(widths)
	for _, row := range rows {
		printRow(row, widths)
	}
}

func printRow(cells []string, widths []int) {
	parts := make([]string, len(widths))
	for i := range widths {
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		visible := VisibleWidth(cell)
		pad := widths[i] - visible
		if pad < 0 {
			pad = 0
		}
		parts[i] = cell + strings.Repeat(" ", pad)
	}
	fmt.Println("  " + strings.Join(parts, "  "))
}

func printSeparator(widths []int) {
	parts := make([]string, len(widths))
	for i, w := range widths {
		parts[i] = strings.Repeat("─", w)
	}
	fmt.Println("  " + strings.Join(parts, "  "))
}

// SanitizeCell strips all ANSI escape sequences and non-printable control
// characters from untrusted API data. Call this on API-sourced strings (names,
// symbols) before passing to PrintTable. Internal color formatting (e.g.
// ColorPercent) operates on numeric values and should NOT be sanitized.
func SanitizeCell(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '\033' {
			// Skip entire escape sequence (ESC [ ... final_byte).
			j := i + 1
			if j < len(runes) && runes[j] == '[' {
				k := j + 1
				for k < len(runes) && runes[k] >= 0x20 && runes[k] <= 0x3F {
					k++
				}
				if k < len(runes) && runes[k] >= 0x40 && runes[k] <= 0x7E {
					i = k
				}
			}
			continue
		}
		if unicode.IsControl(r) && r != '\t' {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

