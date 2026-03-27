package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/mdnmdn/bits/internal/legacy/display"
	"github.com/mdnmdn/bits/internal/legacy/provider"
	"github.com/mdnmdn/bits/internal/legacy/model"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type marketsState int

const (
	marketsLoading marketsState = iota
	marketsLoaded
	marketsDetail
)

type MarketsModel struct {
	client   provider.Provider
	vs       string
	category string
	total    int
	coins    []model.MarketCoin
	cursor   int
	state    marketsState
	detail   DetailModel
	err      error
	width    int
	height   int
}

type coinsLoadedMsg struct {
	coins []model.MarketCoin
	err   error
}

func NewMarketsModel(client provider.Provider, vs, category string, total int) MarketsModel {
	return MarketsModel{
		client:   client,
		vs:       vs,
		category: category,
		total:    total,
		state:    marketsLoading,
	}
}

func (m MarketsModel) Init() tea.Cmd {
	return m.fetchCoins()
}

func (m MarketsModel) fetchCoins() tea.Cmd {
	return func() tea.Msg {
		ml, ok := m.client.(provider.MarketLister)
		if !ok {
			return coinsLoadedMsg{err: model.ErrNotSupported}
		}
		coins, err := ml.FetchAllMarkets(context.Background(), m.vs, m.total, "market_cap_desc", m.category)
		return coinsLoadedMsg{coins: coins, err: err}
	}
}

func (m MarketsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case coinsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		m.coins = msg.coins
		m.state = marketsLoaded
		return m, nil

	case tea.KeyMsg:
		if m.state == marketsDetail {
			updated, cmd := m.detail.Update(msg)
			detail := updated.(DetailModel)
			if detail.Done {
				m.state = marketsLoaded
				return m, nil
			}
			m.detail = detail
			return m, cmd
		}

		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "j", "down":
			if m.cursor < len(m.coins)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "enter":
			if m.state == marketsLoaded && len(m.coins) > 0 {
				coin := m.coins[m.cursor]
				m.detail = NewDetailModel(m.client, coin.ID, m.vs, m.width, m.height)
				m.state = marketsDetail
				return m, m.detail.Init()
			}
		}

	default:
		if m.state == marketsDetail {
			updated, cmd := m.detail.Update(msg)
			m.detail = updated.(DetailModel)
			return m, cmd
		}
	}

	return m, nil
}

func (m MarketsModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	if m.state == marketsLoading {
		return renderLoading(fmt.Sprintf("Fetching top %d coins by market cap...", m.total), m.width, m.height)
	}

	if m.state == marketsDetail {
		return m.detail.View()
	}

	subtitle := fmt.Sprintf("TUI — Top %d by Market Cap", m.total)
	if m.category != "" {
		subtitle += " [" + m.category + "]"
	}

	var b strings.Builder
	b.WriteString(BrandTitle(subtitle))
	b.WriteString("\n\n")

	header := fmt.Sprintf("  %-4s %-20s %-10s %14s %12s %12s %10s",
		"#", "Name", "Symbol", "Price", "Market Cap", "Volume", "24h")
	b.WriteString(HeaderStyle.Render(header))
	b.WriteString("\n")

	// Reserve lines for: title(1) + blank(1) + header(1) + blank(1) + help(1) + border(2) = 7
	visibleRows := m.height - 9
	if visibleRows < 5 {
		visibleRows = 5
	}

	start := 0
	if m.cursor >= visibleRows {
		start = m.cursor - visibleRows + 1
	}
	end := start + visibleRows
	if end > len(m.coins) {
		end = len(m.coins)
	}

	for i := start; i < end; i++ {
		c := m.coins[i]
		// Pad percent to fixed width BEFORE applying color (ANSI codes break fmt width).
		pctStr := fmt.Sprintf("%10s", display.FormatPercent(c.PriceChangePercentage24h))
		pctStr = ColorPercent(c.PriceChangePercentage24h, pctStr)

		row := fmt.Sprintf("%-4d %-20s %-10s %14s %12s %12s %s",
			c.MarketCapRank,
			truncate(display.SanitizeCell(c.Name), 20),
			truncate(display.FormatSymbol(c.Symbol), 10),
			display.FormatPrice(c.CurrentPrice, m.vs),
			display.FormatLargeNumber(c.MarketCap, m.vs),
			display.FormatLargeNumber(c.TotalVolume, m.vs),
			pctStr,
		)
		if i == m.cursor {
			row = SelectedStyle.Render(HighlightSymbol + row)
		} else {
			row = "  " + row
		}
		b.WriteString(row)
		b.WriteString("\n")
	}

	help := HelpStyle.Render(listHelpText)
	content := b.String() + "\n" + help

	return renderFrame(m.width, m.height, content)
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}

func renderLoading(msg string, width, height int) string {
	content := BrandTitle("Loading…") + "\n\n"
	content += lipgloss.Place(
		width-4, height-6,
		lipgloss.Center, lipgloss.Center,
		DimStyle.Render(msg),
	)
	return renderFrame(width, height, content)
}
