package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/coingecko/coingecko-cli/internal/api"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type trendingState int

const (
	trendingLoading trendingState = iota
	trendingLoaded
	trendingDetail
)

type TrendingModel struct {
	client *api.Client
	vs     string
	coins  []api.TrendingCoinWrapper
	cursor int
	state  trendingState
	detail DetailModel
	err    error
	width  int
	height int
}

type trendingLoadedMsg struct {
	resp *api.TrendingResponse
	err  error
}

const defaultTrendingLimit = 30

func NewTrendingModel(client *api.Client, vs string) TrendingModel {
	return TrendingModel{
		client: client,
		vs:     vs,
		state:  trendingLoading,
	}
}

func (m TrendingModel) Init() tea.Cmd {
	return func() tea.Msg {
		resp, err := m.client.SearchTrending(context.Background(), "")
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
		return "Loading trending coins...\n"
	}

	if m.state == trendingDetail {
		return m.detail.View()
	}

	var b strings.Builder
	b.WriteString(TitleStyle.Render("Trending Coins"))
	b.WriteString("\n\n")

	header := fmt.Sprintf("  %-4s %-25s %-8s %15s", "#", "Name", "Symbol", "Market Cap Rank")
	b.WriteString(HeaderStyle.Render(header))
	b.WriteString("\n")

	limit := defaultTrendingLimit
	if len(m.coins) < limit {
		limit = len(m.coins)
	}

	for i := 0; i < limit; i++ {
		c := m.coins[i].Item
		rank := "-"
		if c.MarketCapRank > 0 {
			rank = fmt.Sprintf("%d", c.MarketCapRank)
		}
		row := fmt.Sprintf("  %-4d %-25s %-8s %15s",
			i+1,
			truncate(c.Name, 25),
			c.Symbol,
			rank,
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
