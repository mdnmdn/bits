package display

import (
	"fmt"
	"os"
	"strings"
)

// Brand color: CoinGecko green #4BCC00 → RGB(75, 204, 0)
const (
	brandGreen = "\033[38;2;75;204;0m"
	dimColor   = "\033[2m"
	cyanColor  = "\033[36m"
	yellowBold = "\033[1;33m"
	boxWidth   = 78 // inner width of the welcome box
)

var asciiLogo = []string{
	"  ██████╗ ██████╗ ██╗███╗   ██╗ ██████╗ ███████╗ ██████╗██╗  ██╗ ██████╗ ",
	" ██╔════╝██╔═══██╗██║████╗  ██║██╔════╝ ██╔════╝██╔════╝██║ ██╔╝██╔═══██╗",
	" ██║     ██║   ██║██║██╔██╗ ██║██║  ███╗█████╗  ██║     █████╔╝ ██║   ██║",
	" ██║     ██║   ██║██║██║╚██╗██║██║   ██║██╔══╝  ██║     ██╔═██╗ ██║   ██║",
	" ╚██████╗╚██████╔╝██║██║ ╚████║╚██████╔╝███████╗╚██████╗██║  ██╗╚██████╔╝",
	"  ╚═════╝ ╚═════╝ ╚═╝╚═╝  ╚═══╝ ╚═════╝ ╚══════╝ ╚═════╝╚═╝  ╚═╝ ╚═════╝ ",
}

// PrintLogo prints the full ASCII art CoinGecko logo in brand green to stderr.
func PrintLogo() {
	if !colorEnabled() {
		for _, line := range asciiLogo {
			fmt.Fprintln(os.Stderr, line)
		}
		fmt.Fprintln(os.Stderr)
		return
	}
	fmt.Fprintln(os.Stderr)
	for _, line := range asciiLogo {
		fmt.Fprintf(os.Stderr, "%s%s%s\n", brandGreen, line, colorReset)
	}
	fmt.Fprintln(os.Stderr)
}

// PrintWelcomeBox prints a bordered quick-start box to stderr.
func PrintWelcomeBox() {
	w := os.Stderr
	top := "+" + strings.Repeat("-", boxWidth) + "+"
	blank := "|" + strings.Repeat(" ", boxWidth) + "|"
	sep := boxRow(w, dimColor+strings.Repeat("-", boxWidth-2)+colorReset, boxWidth-2)

	fmt.Fprintln(w, top)
	fmt.Fprintln(w, blank)
	printColoredRow(w, yellowBold+"Official API Command Line Interface"+colorReset, 36)
	fmt.Fprintln(w, blank)
	fmt.Fprintln(w, sep)
	fmt.Fprintln(w, blank)
	printPlainRow(w, "  Quick Start")
	fmt.Fprintln(w, blank)
	printCmdRow(w, "cg auth", "# Set up your API key")
	printCmdRow(w, "cg price --ids bitcoin", "# Get BTC price")
	printCmdRow(w, "cg markets --total 100", "# Top 100 by mkt cap")
	printCmdRow(w, "cg search ethereum", "# Search for a coin")
	printCmdRow(w, "cg trending", "# Trending coins")
	printCmdRow(w, "cg history bitcoin --days 30", "# 30-day OHLC history")
	printCmdRow(w, "cg top-gainers-losers", "# Top gainers (paid)")
	printCmdRow(w, "cg tui markets", "# Interactive TUI")
	fmt.Fprintln(w, blank)
	fmt.Fprintln(w, sep)
	fmt.Fprintln(w, blank)
	printColoredRow(w, "  "+dimColor+"Docs: "+colorReset+cyanColor+"https://docs.coingecko.com"+colorReset, 34)
	fmt.Fprintln(w, blank)
	fmt.Fprintln(w, top)
	fmt.Fprintln(w)
}

func printPlainRow(w *os.File, text string) {
	pad := boxWidth - 2 - len(text)
	if pad < 0 {
		pad = 0
	}
	fmt.Fprintf(w, "| %s%s |\n", text, strings.Repeat(" ", pad))
}

func printColoredRow(w *os.File, content string, visible int) {
	pad := boxWidth - 2 - visible
	if pad < 0 {
		pad = 0
	}
	if !colorEnabled() {
		plain := ansiRegex.ReplaceAllString(content, "")
		plainPad := boxWidth - 2 - len(plain)
		if plainPad < 0 {
			plainPad = 0
		}
		fmt.Fprintf(w, "| %s%s |\n", plain, strings.Repeat(" ", plainPad))
		return
	}
	fmt.Fprintf(w, "| %s%s |\n", content, strings.Repeat(" ", pad))
}

func printCmdRow(w *os.File, cmd, comment string) {
	// Layout: "| " + "  " + "$" + " " + cmd(30) + " " + comment + pad + " |"
	cmdPad := 30 - len(cmd)
	if cmdPad < 0 {
		cmdPad = 0
	}
	commentPad := 41 - len(comment)
	if commentPad < 0 {
		commentPad = 0
	}
	if colorEnabled() {
		fmt.Fprintf(w, "|   %s$%s %s%s %s%s%s |\n",
			brandGreen, colorReset,
			cmd, strings.Repeat(" ", cmdPad),
			dimColor, comment, colorReset+strings.Repeat(" ", commentPad))
	} else {
		fmt.Fprintf(w, "|   $ %s%s %s%s |\n",
			cmd, strings.Repeat(" ", cmdPad),
			comment, strings.Repeat(" ", commentPad))
	}
}

func boxRow(w *os.File, content string, visible int) string {
	pad := boxWidth - 2 - visible
	if pad < 0 {
		pad = 0
	}
	return fmt.Sprintf("| %s%s |", content, strings.Repeat(" ", pad))
}

// PrintBanner prints a compact one-line CoinGecko banner to stderr.
// Writing to stderr keeps stdout clean for piped data.
func PrintBanner() {
	if !colorEnabled() {
		fmt.Fprint(os.Stderr, "\n  CoinGecko CLI  —  Real-time crypto data\n\n")
		return
	}
	fmt.Fprintf(os.Stderr, "\n  %s◆ CoinGecko%s %sCLI  —  Real-time crypto data%s\n\n",
		brandGreen, colorReset, dimColor, colorReset)
}
