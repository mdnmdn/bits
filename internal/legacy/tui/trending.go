package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/mdnmdn/bits/internal/legacy/display"
	"github.com/mdnmdn/bits/internal/legacy/provider"
	"github.com/mdnmdn/bits/internal/legacy/model"

	tea "github.com/charmbracelet/bubbletea"
)

type trendingState int

const (
	trendingLoading trendingState = iota
	trendingLoaded
	trendingDetail
)

type TrendingModel struct {
	client  provider.Provider
	vs      string
	showMax string
	limit   int
	coins   []model.TrendingCoinWrapper
	cursor  int
	state   trendingState
	detail  DetailModel
	err     error
	width   int
	height  int
}

type trendingLoadedMsg struct {
	resp *model.TrendingResponse
	err  error
}

func NewTrendingModel(client provider.Provider, vs, showMax string) TrendingModel {
	limit := 15
	if showMax != "" {
		limit = 30
	}
	return TrendingModel{
		client:  client,
		vs:      vs,
		showMax: showMax,
		limit:   limit,
		state:   trendingLoading,
	}
}

func (m TrendingModel) Init() tea.Cmd {
	return func() tea.Msg {
		tp, ok := m.client.(provider.TrendingProvider)
		if !ok {
			return trendingLoadedMsg{err: model.ErrNotSupported}
		}
		resp, err := tp.SearchTrending(context.Background(), m.showMax)
		return trendingLoadedMsg{resp: resp, err: err}
	}
}

func (m TrendingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case trendingLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		m.coins = msg.resp.Coins
		m.state = trendingLoaded
		return m, nil

	case tea.KeyMsg:
		if m.state == trendingDetail {
			updated, cmd := m.detail.Update(msg)
			detail := updated.(DetailModel)
			if detail.Done {
				m.state = trendingLoaded
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
			if m.state == trendingLoaded && len(m.coins) > 0 {
				coin := m.coins[m.cursor].Item
				m.detail = NewDetailModel(m.client, coin.ID, m.vs, m.width, m.height)
				m.state = trendingDetail
				return m, m.detail.Init()
			}
		}

	default:
		if m.state == trendingDetail {
			updated, cmd := m.detail.Update(msg)
			m.detail = updated.(DetailModel)
			return m, cmd
		}
	}

	return m, nil
}

func (m TrendingModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	if m.state == trendingLoading {
		return renderLoading("Fetching trending coins...", m.width, m.height)
	}

	if m.state == trendingDetail {
		return m.detail.View()
	}

	var b strings.Builder
	b.WriteString(BrandTitle(fmt.Sprintf("TUI — Top %d Trending Coins (24h)", m.limit)))
	b.WriteString("\n\n")

	header := fmt.Sprintf("  %-6s %-7s %-20s %-8s %14s %10s",
		"Trend", "MCap #", "Name", "Symbol", "Price (USD)", "24h")
	b.WriteString(HeaderStyle.Render(header))
	b.WriteString("\n")

	limit := m.limit
	if len(m.coins) < limit {
		limit = len(m.coins)
	}

	for i := 0; i < limit; i++ {
		c := m.coins[i].Item

		trendRank := fmt.Sprintf("#%d", i+1)
		mcapRank := "—"
		if c.MarketCapRank > 0 {
			mcapRank = fmt.Sprintf("#%d", c.MarketCapRank)
		}

		priceStr := fmt.Sprintf("%14s", "—")
		changeStr := fmt.Sprintf("%10s", "—")
		if c.Data != nil {
			priceStr = fmt.Sprintf("%14s", display.FormatPrice(c.Data.Price, "usd"))
			if pct, ok := c.Data.PriceChangePercentage24h["usd"]; ok {
				changeStr = fmt.Sprintf("%10s", display.FormatPercent(pct))
				changeStr = ColorPercent(pct, changeStr)
			}
		}

		row := fmt.Sprintf("%-6s %-7s %-20s %-8s %s %s",
			trendRank,
			mcapRank,
			truncate(display.SanitizeCell(c.Name), 20),
			display.SanitizeCell(c.Symbol),
			priceStr,
			changeStr,
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
