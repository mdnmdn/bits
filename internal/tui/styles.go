package tui

import "github.com/charmbracelet/lipgloss"

var (
	HeaderStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229"))
	SelectedStyle = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("236"))
	HelpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	GreenStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	RedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	DimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	TitleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99")).MarginBottom(1)
)

func ColorPercent(pct float64, s string) string {
	if pct > 0 {
		return GreenStyle.Render(s)
	} else if pct < 0 {
		return RedStyle.Render(s)
	}
	return s
}
