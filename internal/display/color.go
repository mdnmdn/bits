package display

import "fmt"

const (
	colorReset = "\033[0m"
	colorGreen = "\033[32m"
	colorRed   = "\033[31m"
)

func ColorPercent(pct float64) string {
	s := FormatPercent(pct)
	if pct > 0 {
		return fmt.Sprintf("%s%s%s", colorGreen, s, colorReset)
	} else if pct < 0 {
		return fmt.Sprintf("%s%s%s", colorRed, s, colorReset)
	}
	return s
}
