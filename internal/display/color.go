package display

import (
	"fmt"
	"os"
	"regexp"
	"sync"

	"golang.org/x/term"
)

const (
	colorReset = "\033[0m"
	colorGreen = "\033[32m"
	colorRed   = "\033[31m"
)

var (
	colorOnce    sync.Once
	colorAllowed bool
	ansiRegex    = regexp.MustCompile(`\x1b\[[0-9;]*m`)
)

func colorEnabled() bool {
	colorOnce.Do(func() {
		if os.Getenv("NO_COLOR") != "" {
			colorAllowed = false
			return
		}
		colorAllowed = term.IsTerminal(int(os.Stdout.Fd()))
	})
	return colorAllowed
}

// VisibleWidth returns the display width of a string after stripping ANSI escapes.
func VisibleWidth(s string) int {
	return len(ansiRegex.ReplaceAllString(s, ""))
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
