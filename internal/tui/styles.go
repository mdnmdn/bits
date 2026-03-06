package tui

import "github.com/charmbracelet/lipgloss"

// Brand colors matching CoinGecko identity.
var (
	GeckoGreen = lipgloss.Color("#8CC351")
	Gold       = lipgloss.Color("#FFD700")
)

var (
	HeaderStyle   = lipgloss.NewStyle().Bold(true).Foreground(Gold)
	SelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")).Background(GeckoGreen)
	HelpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	GreenStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	RedStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	DimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	TitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(GeckoGreen)
	LabelStyle    = lipgloss.NewStyle().Bold(true).Foreground(Gold)
)

const HighlightSymbol = "▶ "

const listHelpText = "  ↑/k  ↓/j  navigate    Enter  detail    q/Esc  quit"

func ColorPercent(pct float64, s string) string {
	if pct > 0 {
		return GreenStyle.Render(s)
	} else if pct < 0 {
		return RedStyle.Render(s)
	}
	return s
}

// BrandTitle returns the branded title line: ◆ CoinGecko <subtitle>
func BrandTitle(subtitle string) string {
	return TitleStyle.Render(" ◆ CoinGecko") + " " + DimStyle.Render(subtitle)
}

// FrameStyle returns a bordered lipgloss style for the TUI outer frame.
func FrameStyle(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(GeckoGreen).
		Width(width - 2).
		Height(height - 2).
		PaddingLeft(1).
		PaddingRight(1)
}

// renderFrame wraps content in a branded frame and places it in the terminal.
func renderFrame(width, height int, content string) string {
	frame := FrameStyle(width, height)
	return lipgloss.Place(width, height, lipgloss.Left, lipgloss.Top, frame.Render(content))
}

// renderPlaceholder renders a branded frame with a title and message body.
func renderPlaceholder(width, height int, title, body string) string {
	content := BrandTitle(title) + "\n\n" + body + "\n\nPress esc to go back."
	return renderFrame(width, height, content)
}
