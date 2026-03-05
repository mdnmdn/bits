package display

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

const (
	colorReset = "\033[0m"
	colorGreen = "\033[32m"
	colorRed   = "\033[31m"
)

func colorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func ColorPercent(pct float64) string {
	s := FormatPercent(pct)
	if !colorEnabled() {
		return s
	}
	if pct > 0 {
		return fmt.Sprintf("%s%s%s", colorGreen, s, colorReset)
	} else if pct < 0 {
		return fmt.Sprintf("%s%s%s", colorRed, s, colorReset)
	}
	return s
}
