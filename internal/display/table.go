package display

import (
	"fmt"
	"strings"
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
