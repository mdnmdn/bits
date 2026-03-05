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
			cell = sanitizeCell(cells[i])
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

// sanitizeCell strips non-printable control characters from a string while
// preserving ANSI color escape sequences (ESC[...m) that we generate ourselves.
func sanitizeCell(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '\033' {
			// Allow our ANSI SGR sequences (ESC[...m), pass them through.
			j := i + 1
			if j < len(runes) && runes[j] == '[' {
				k := j + 1
				for k < len(runes) && ((runes[k] >= '0' && runes[k] <= '9') || runes[k] == ';') {
					k++
				}
				if k < len(runes) && runes[k] == 'm' {
					// Valid SGR sequence — copy it through.
					for x := i; x <= k; x++ {
						b.WriteRune(runes[x])
					}
					i = k
					continue
				}
			}
			// Non-SGR escape sequence — strip it.
			continue
		}
		if unicode.IsControl(r) && r != '\t' {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
