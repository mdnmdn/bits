package display

import (
	"fmt"
	"os"
	"regexp"
	"sync"

	"golang.org/x/term"
)

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorHeader = "\033[1;38;2;255;232;102m" // bold #FFE866
)

var (
	colorOnce    sync.Once
	colorAllowed bool
	ansiRegex    = regexp.MustCompile(`\x1b\[[0-9;]*m`)
)

// ColorEnabled reports whether color output is allowed (TTY and NO_COLOR not set).
func ColorEnabled() bool {
	colorOnce.Do(func() {
		if os.Getenv("NO_COLOR") != "" {
			colorAllowed = false
			return
		}
		// Check stderr since banner/logo write there; table data goes to stdout
		// but is never colored when piped (colorEnabled gates all ANSI output).
		colorAllowed = term.IsTerminal(int(os.Stderr.Fd()))
	})
	return colorAllowed
}

// StderrIsTerminal reports whether stderr is connected to an interactive terminal.
func StderrIsTerminal() bool {
	return term.IsTerminal(int(os.Stderr.Fd()))
}

// VisibleWidth returns the display width of a string after stripping ANSI escapes.
func VisibleWidth(s string) int {
	return len(ansiRegex.ReplaceAllString(s, ""))
}

// ColorHeader wraps s in bold #FFE866 (gold) when color output is enabled.
func ColorHeader(s string) string {
	if !ColorEnabled() {
		return s
	}
	return colorHeader + s + colorReset
}

func ColorPercent(pct float64) string {
	s := FormatPercent(pct)
	if !ColorEnabled() {
		return s
	}
	if pct > 0 {
		return fmt.Sprintf("%s%s%s", colorGreen, s, colorReset)
	} else if pct < 0 {
		return fmt.Sprintf("%s%s%s", colorRed, s, colorReset)
	}
	return s
}
