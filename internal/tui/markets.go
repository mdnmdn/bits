package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/coingecko/coingecko-cli/internal/api"
	"github.com/coingecko/coingecko-cli/internal/display"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const defaultMarketsLimit = 50

type marketsState int

const (
	marketsLoading marketsState = iota
	marketsLoaded
	marketsDetail
)

type MarketsModel struct {
	client   *api.Client
	vs       string
	category string
	coins    []api.MarketCoin
	cursor   int
	state    marketsState
	detail   DetailModel
	err      error
	width    int
	height   int
}

type coinsLoadedMsg struct {
	coins []api.MarketCoin
	err   error
}

func NewMarketsModel(client *api.Client, vs, category string) MarketsModel {
	return MarketsModel{
		client:   client,
		vs:       vs,
		category: category,
		state:    marketsLoading,
	}
}

func (m MarketsModel) Init() tea.Cmd {
	return m.fetchCoins()
}

func (m MarketsModel) fetchCoins() tea.Cmd {
	return func() tea.Msg {
		coins, err := m.client.CoinMarkets(context.Background(), m.vs, defaultMarketsLimit, 1, "market_cap_desc", m.category)
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
		case "q", "ctrl+c":
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
		return "Loading markets...\n"
	}

	if m.state == marketsDetail {
		return m.detail.View()
	}

	var b strings.Builder
	b.WriteString(TitleStyle.Render(fmt.Sprintf("Markets — Top %d by Market Cap", defaultMarketsLimit)))
	b.WriteString("\n\n")

	header := fmt.Sprintf("  %-4s %-20s %-6s %12s %12s %10s", "#", "Name", "Symbol", "Price", "Market Cap", "24h %")
	b.WriteString(HeaderStyle.Render(header))
	b.WriteString("\n")

	visibleRows := m.height - 6
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
		pctStr := display.FormatPercent(c.PriceChangePercentage24h)
		pctStr = ColorPercent(c.PriceChangePercentage24h, pctStr)

		row := fmt.Sprintf("  %-4d %-20s %-6s %12s %12s %10s",
			c.MarketCapRank,
			truncate(display.SanitizeCell(c.Name), 20),
			display.FormatSymbol(c.Symbol),
			display.FormatPrice(c.CurrentPrice, m.vs),
			display.FormatLargeNumber(c.MarketCap, m.vs),
			pctStr,
		)
		if i == m.cursor {
			row = SelectedStyle.Render(row)
		}
		b.WriteString(row)
		b.WriteString("\n")
	}

	help := HelpStyle.Render("j/k: navigate • enter: detail • q: quit")
	b.WriteString("\n")
	b.WriteString(help)

	return lipgloss.NewStyle().MaxWidth(m.width).Render(b.String())
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}
