package display

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

const (
	brandGreen = "\033[38;2;75;204;0m"
	dimColor   = "\033[2m"
	cyanColor  = "\033[36m"
	boxWidth   = 78
)

var asciiLogo = []string{
	"  ██████╗ ██████╗ ██╗███╗   ██╗ ██████╗ ███████╗ ██████╗██╗  ██╗ ██████╗ ",
	" ██╔════╝██╔═══██╗██║████╗  ██║██╔════╝ ██╔════╝██╔════╝██║ ██╔╝██╔═══██╗",
	" ██║     ██║   ██║██║██╔██╗ ██║██║  ███╗█████╗  ██║     █████╔╝ ██║   ██║",
	" ██║     ██║   ██║██║██║╚██╗██║██║   ██║██╔══╝  ██║     ██╔═██╗ ██║   ██║",
	" ╚██████╗╚██████╔╝██║██║ ╚████║╚██████╔╝███████╗╚██████╗██║  ██╗╚██████╔╝",
	"  ╚═════╝ ╚═════╝ ╚═╝╚═╝  ╚═══╝ ╚═════╝ ╚══════╝ ╚═════╝╚═╝  ╚═╝ ╚═════╝ ",
}

func PrintLogo() {
	if !ColorEnabled() {
		for _, line := range asciiLogo {
			_, _ = fmt.Fprintln(os.Stderr, line)
		}
		_, _ = fmt.Fprintln(os.Stderr)
		return
	}
	_, _ = fmt.Fprintln(os.Stderr)
	for _, line := range asciiLogo {
		_, _ = fmt.Fprintf(os.Stderr, "%s%s%s\n", brandGreen, line, colorReset)
	}
	_, _ = fmt.Fprintln(os.Stderr)
}

const BannerLines = 3

func FprintBanner(w io.Writer) {
	colored := false
	if os.Getenv("NO_COLOR") == "" {
		if f, ok := w.(*os.File); ok {
			colored = term.IsTerminal(int(f.Fd()))
		}
	}
	if !colored {
		_, _ = fmt.Fprint(w, "\n  bits  —  crypto data CLI\n\n")
		return
	}
	_, _ = fmt.Fprintf(w, "\n  %s◆ bits%s %s—  crypto data CLI%s\n\n",
		brandGreen, colorReset, dimColor, colorReset)
}

func PrintBanner() {
	FprintBanner(os.Stderr)
}
