package display

import (
	"fmt"
	"os"
)

// Brand color: CoinGecko green #4BCC00 → RGB(75, 204, 0)
const (
	brandGreen = "\033[38;2;75;204;0m"
	dimColor   = "\033[2m"
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
