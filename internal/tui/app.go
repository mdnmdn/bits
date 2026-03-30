package tui

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mdnmdn/bits/internal/tui/section"
	"github.com/mdnmdn/bits/pkg/capability"
	"github.com/mdnmdn/bits/pkg/model"
)

type Options struct {
	Section         string
	Provider        string
	Market          string
	Symbol          string
	RefreshInterval string
}

type App struct {
	opts            Options
	width           int
	height          int
	section         SectionModel
	tp              *TUIProvider
	ctx             context.Context
	err             error
	refreshInterval time.Duration
	refreshEnabled  bool
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED"))

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF"))

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#3B82F6"))

	providerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981"))

	marketStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B"))
)

func NewApp(opts Options) *App {
	tp, err := NewTUIProvider()
	if err != nil {
		return &App{opts: opts, err: err}
	}

	ctx := context.Background()
	sectionName := Section(opts.Section)
	if sectionName == "" {
		sectionName = SectionPrices
	}

	refreshInterval := ParseRefreshInterval(opts.RefreshInterval)

	var sect SectionModel
	var wrapper ProviderWrapper = tp
	switch sectionName {
	case SectionPrices:
		sect = section.NewPricesModel(ctx, wrapper, opts.Provider, model.MarketType(opts.Market), opts.Symbol)
	default:
		sect = section.NewPricesModel(ctx, wrapper, opts.Provider, model.MarketType(opts.Market), opts.Symbol)
	}

	app := &App{
		opts:            opts,
		section:         sect,
		tp:              tp,
		ctx:             ctx,
		refreshInterval: refreshInterval,
		refreshEnabled:  refreshInterval > 0,
	}

	return app
}

func (a *App) Init() tea.Cmd {
	if a.section != nil {
		return a.section.Init()
	}
	return nil
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		if a.section != nil {
			_, cmd := a.section.Update(msg)
			return a, cmd
		}
		return a, nil

	case tea.KeyMsg:
		if a.section != nil {
			updated, cmd := a.section.Update(msg)
			a.section = updated
			if cmd != nil {
				return a, cmd
			}
		}

		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return a, tea.Quit
		case "p":
			return a, a.cycleProvider()
		case "m":
			return a, a.cycleMarket()
		case "r":
			return a, a.cycleRefresh()
		}
	}
	return a, nil
}

func (a *App) View() string {
	if a.err != nil {
		return renderError(a.err.Error(), a.width, a.height)
	}

	if a.width == 0 || a.height == 0 {
		return "Loading..."
	}

	section := a.opts.Section
	if section == "" {
		section = "prices"
	}

	header := titleStyle.Render("bits — TUI") + " │ " + sectionStyle.Render(section)
	if a.opts.Provider != "" {
		header += " │ " + providerStyle.Render(a.opts.Provider)
	}
	if a.opts.Market != "" {
		header += " │ " + marketStyle.Render(a.opts.Market)
	}
	if a.refreshEnabled {
		header += " │ " + lipgloss.NewStyle().Foreground(lipgloss.Color("#8B5CF6")).Render("↻ "+a.refreshInterval.String())
	}

	var content string
	if a.section != nil {
		content = a.section.View()
	} else {
		content = fmt.Sprintf(`%s


%s`,
			header,
			subtitleStyle.Render("Press 'q' to quit"),
		)
	}

	return lipgloss.Place(a.width, a.height, lipgloss.Top, lipgloss.Left, content)
}

func (a *App) cycleProvider() tea.Cmd {
	feature := a.currentFeature()
	if feature == "" {
		return nil
	}

	providers := a.tp.GetAvailableProviders(feature)
	if len(providers) == 0 {
		return nil
	}

	currentIdx := -1
	for i, p := range providers {
		if p == a.opts.Provider {
			currentIdx = i
			break
		}
	}

	nextIdx := (currentIdx + 1) % len(providers)
	a.opts.Provider = providers[nextIdx]

	if pm, ok := a.section.(interface{ SetProvider(string) }); ok {
		pm.SetProvider(a.opts.Provider)
	}

	return nil
}

func (a *App) cycleMarket() tea.Cmd {
	feature := a.currentFeature()
	if feature == "" {
		return nil
	}

	markets := a.tp.GetAvailableMarkets(a.opts.Provider, feature)
	if len(markets) == 0 {
		return nil
	}

	currentIdx := -1
	for i, m := range markets {
		if string(m) == a.opts.Market {
			currentIdx = i
			break
		}
	}

	nextIdx := (currentIdx + 1) % len(markets)
	a.opts.Market = string(markets[nextIdx])

	if pm, ok := a.section.(interface{ SetMarket(model.MarketType) }); ok {
		pm.SetMarket(model.MarketType(a.opts.Market))
	}

	return nil
}

func (a *App) cycleRefresh() tea.Cmd {
	intervals := []time.Duration{5 * time.Second, 10 * time.Second, 30 * time.Second, 1 * time.Minute, 0}

	currentIdx := 0
	for i, interval := range intervals {
		if a.refreshInterval == interval {
			currentIdx = i
			break
		}
	}

	nextIdx := (currentIdx + 1) % len(intervals)
	a.refreshInterval = intervals[nextIdx]
	a.refreshEnabled = a.refreshInterval > 0

	return nil
}

func (a *App) currentFeature() capability.Feature {
	sectionName := Section(a.opts.Section)
	if sectionName == "" {
		sectionName = SectionPrices
	}
	return capability.Feature(sectionName.Feature())
}

func renderError(err string, width, height int) string {
	content := titleStyle.Render("bits — TUI\n\n")
	content += lipgloss.Place(width-4, height-4, lipgloss.Center, lipgloss.Center,
		lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Render("Error: "+err))
	return content
}

func Main(opts Options) error {
	p := tea.NewProgram(NewApp(opts), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
