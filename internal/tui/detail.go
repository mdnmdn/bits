package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

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
		return renderLoading(fmt.Sprintf("Fetching 7-day chart for %s...", m.coinID), m.width, m.height)
	}
	if m.err != nil {
		return renderPlaceholder(m.width, m.height, "Error", fmt.Sprintf("Error: %v", m.err))
	}
	if m.coin == nil {
		return renderPlaceholder(m.width, m.height, "Detail", "No data available.")
	}

	md := m.coin.MarketData
	vs := m.vs
	coinName := display.SanitizeCell(m.coin.Name)
	coinSymbol := display.FormatSymbol(m.coin.Symbol)

	leftWidth := 35
	if m.width > 0 {
		leftWidth = m.width * 30 / 100
		if leftWidth < 30 {
			leftWidth = 30
		}
	}

	// Left panel: info
	var left strings.Builder
	infoTitle := LabelStyle.Render(" Info ")
	infoBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(GeckoGreen).
		Width(leftWidth - 2).
		PaddingLeft(1)

	left.WriteString("\n")
	if md != nil {
		addDetailField(&left, "Rank", fmt.Sprintf("%d", m.coin.MarketCapRank))
		addDetailField(&left, "Name", coinName)
		addDetailField(&left, "Symbol", coinSymbol)
		addDetailField(&left, "ID", display.SanitizeCell(m.coin.ID))
		left.WriteString("\n")
		addDetailField(&left, "Price", display.FormatPrice(md.CurrentPrice[vs], vs))
		addDetailField(&left, "Mkt Cap", display.FormatLargeNumber(md.MarketCap[vs], vs))
		addDetailField(&left, "Vol 24h", display.FormatLargeNumber(md.TotalVolume[vs], vs))
		addDetailField(&left, "24h Chg", display.ColorPercent(md.PriceChangePercentage24h))
		left.WriteString("\n")
		addDetailField(&left, "Hi 24h", display.FormatPrice(md.High24h[vs], vs))
		addDetailField(&left, "Lo 24h", display.FormatPrice(md.Low24h[vs], vs))
		left.WriteString("\n")
		addDetailField(&left, "ATH", display.FormatPrice(md.ATH[vs], vs))
		addDimField(&left, " date", formatATHDate(md.ATHDate[vs]))
		addDetailField(&left, " chg%", ColorPercent(md.ATHChangePercentage[vs], display.FormatPercent(md.ATHChangePercentage[vs])))
		addDetailField(&left, "ATL", display.FormatPrice(md.ATL[vs], vs))
		addDimField(&left, " date", formatATHDate(md.ATLDate[vs]))
		addDetailField(&left, " chg%", ColorPercent(md.ATLChangePercentage[vs], display.FormatPercent(md.ATLChangePercentage[vs])))
		left.WriteString("\n")
		addDetailField(&left, "Circulating", display.FormatSupply(md.CirculatingSupply))
		if md.TotalSupply > 0 {
			addDetailField(&left, "Total Supply", display.FormatSupply(md.TotalSupply))
		}
	}
	left.WriteString("\n")

	leftPanel := infoBox.SetString(infoTitle).Render(left.String())

	// Right panel: chart
	var right strings.Builder
	chartTitle := LabelStyle.Render(" 7-Day Price (USD) ")
	chartBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(GeckoGreen).
		Width(m.width - leftWidth - 8).
		PaddingLeft(1)

	if len(m.ohlc) > 0 {
		chartWidth := m.width - leftWidth - 12
		if chartWidth < 20 {
			chartWidth = 40
		}
		chartHeight := m.height - 10
		if chartHeight < 8 {
			chartHeight = 12
		}
		right.WriteString("\n")
		right.WriteString(renderBrailleChart(m.ohlc, chartWidth, chartHeight, vs))
		right.WriteString("\n")
	} else {
		right.WriteString("\n")
		right.WriteString(DimStyle.Render("  No chart data available"))
		right.WriteString("\n")
	}

	rightPanel := chartBox.SetString(chartTitle).Render(right.String())

	content := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, "  ", rightPanel)
	help := HelpStyle.Render("  Esc / q / ← back to list")

	subtitle := fmt.Sprintf("%s (%s) — Detail", coinName, coinSymbol)
	inner := BrandTitle(subtitle) + "\n\n" + content + "\n\n" + help
	return renderFrame(m.width, m.height, inner)
}

func formatATHDate(s string) string {
	if s == "" {
		return "—"
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.UTC().Format("2006-01-02")
}

func addDetailField(b *strings.Builder, label, value string) {
	fmt.Fprintf(b, " %-12s %s\n", LabelStyle.Render(label), value)
}

func addDimField(b *strings.Builder, label, value string) {
	fmt.Fprintf(b, " %-12s %s\n", DimStyle.Render(label), value)
}
