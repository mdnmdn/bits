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

type DetailModel struct {
	client  *api.Client
	coinID  string
	vs      string
	coin    *api.CoinDetail
	ohlc    api.OHLCData
	loading int // count of pending fetches
	Done    bool
	err     error
	width   int
	height  int
}

type coinDetailMsg struct {
	coin *api.CoinDetail
	err  error
}

type ohlcMsg struct {
	data api.OHLCData
	err  error
}

func NewDetailModel(client *api.Client, coinID, vs string, width, height int) DetailModel {
	return DetailModel{
		client:  client,
		coinID:  coinID,
		vs:      vs,
		loading: 2,
		width:   width,
		height:  height,
	}
}

func (m DetailModel) Init() tea.Cmd {
	return tea.Batch(
		m.fetchDetail(),
		m.fetchOHLC(),
	)
}

func (m DetailModel) fetchDetail() tea.Cmd {
	return func() tea.Msg {
		coin, err := m.client.CoinDetail(context.Background(), m.coinID)
		return coinDetailMsg{coin: coin, err: err}
	}
}

func (m DetailModel) fetchOHLC() tea.Cmd {
	return func() tea.Msg {
		data, err := m.client.CoinOHLC(context.Background(), m.coinID, m.vs, "7", "")
		return ohlcMsg{data: data, err: err}
	}
}

func (m DetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case coinDetailMsg:
		m.loading--
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.coin = msg.coin
		}

	case ohlcMsg:
		m.loading--
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.ohlc = msg.data
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace":
			m.Done = true
			return m, nil
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m DetailModel) View() string {
	if m.loading > 0 {
		return fmt.Sprintf("Loading %s details...\n", m.coinID)
	}
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress esc to go back.\n", m.err)
	}
	if m.coin == nil {
		return "No data available.\n\nPress esc to go back.\n"
	}

	md := m.coin.MarketData
	vs := m.vs

	leftWidth := 35
	if m.width > 0 {
		leftWidth = m.width * 30 / 100
		if leftWidth < 30 {
			leftWidth = 30
		}
	}

	var left strings.Builder
	left.WriteString(TitleStyle.Render(fmt.Sprintf("%s (%s)", display.SanitizeCell(m.coin.Name), strings.ToUpper(display.SanitizeCell(m.coin.Symbol)))))
	left.WriteString("\n\n")

	if md != nil {
		addField(&left, "Price", display.FormatPrice(md.CurrentPrice[vs]))
		addField(&left, "24h Change", display.ColorPercent(md.PriceChangePercentage24h))
		addField(&left, "24h High", display.FormatPrice(md.High24h[vs]))
		addField(&left, "24h Low", display.FormatPrice(md.Low24h[vs]))
		addField(&left, "Market Cap", display.FormatLargeNumber(md.MarketCap[vs]))
		addField(&left, "Volume", display.FormatLargeNumber(md.TotalVolume[vs]))
		addField(&left, "ATH", display.FormatPrice(md.ATH[vs]))
		addField(&left, "ATH Change", display.FormatPercent(md.ATHChangePercentage[vs]))
		addField(&left, "ATL", display.FormatPrice(md.ATL[vs]))
		addField(&left, "ATL Change", display.FormatPercent(md.ATLChangePercentage[vs]))
		addField(&left, "Circulating", display.FormatSupply(md.CirculatingSupply))
		if md.TotalSupply > 0 {
			addField(&left, "Total Supply", display.FormatSupply(md.TotalSupply))
		}
	}

	// Right panel: braille chart
	var right strings.Builder
	right.WriteString(HeaderStyle.Render("7-Day Price Chart"))
	right.WriteString("\n\n")

	if len(m.ohlc) > 0 {
		chartWidth := m.width - leftWidth - 6
		if chartWidth < 20 {
			chartWidth = 40
		}
		chartHeight := m.height - 8
		if chartHeight < 8 {
			chartHeight = 12
		}
		right.WriteString(renderBrailleChart(m.ohlc, chartWidth, chartHeight))
	} else {
		right.WriteString(DimStyle.Render("No chart data available"))
	}

	leftPanel := lipgloss.NewStyle().Width(leftWidth).Render(left.String())
	rightPanel := lipgloss.NewStyle().Width(m.width - leftWidth - 4).Render(right.String())
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, "  ", rightPanel)

	help := HelpStyle.Render("esc: back • q: quit")
	return content + "\n\n" + help
}

func addField(b *strings.Builder, label, value string) {
	b.WriteString(fmt.Sprintf("%-14s %s\n", DimStyle.Render(label), value))
}
