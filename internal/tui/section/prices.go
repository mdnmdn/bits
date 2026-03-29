package section

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mdnmdn/bits/pkg/capability"
	"github.com/mdnmdn/bits/pkg/model"
	"github.com/mdnmdn/bits/pkg/provider"
)

type PriceProviderWrapper interface {
	GetProviderForFeature(ctx context.Context, feature capability.Feature, providerID string, market model.MarketType) (provider.Provider, model.MarketType, error)
}

type PricesModel struct {
	provider   PriceProviderWrapper
	ctx        context.Context
	items      []model.CoinPrice
	cursor     int
	loading    bool
	err        error
	width      int
	height     int
	providerID string
	market     model.MarketType
	symbol     string
}

func NewPricesModel(ctx context.Context, tp PriceProviderWrapper, providerID string, market model.MarketType, symbol string) *PricesModel {
	return &PricesModel{
		provider:   tp,
		ctx:        ctx,
		loading:    true,
		providerID: providerID,
		market:     market,
		symbol:     symbol,
	}
}

func (m *PricesModel) Init() tea.Cmd {
	return func() tea.Msg {
		m.fetchPrices()
		return nil
	}
}

func (m *PricesModel) fetchPrices() {
	feature := capability.FeaturePrice

	var ids []string
	if m.symbol != "" {
		ids = strings.Split(m.symbol, ",")
	} else {
		ids = []string{"bitcoin", "ethereum", "solana", "binancecoin", "tether"}
	}

	p, actualMarket, err := m.provider.GetProviderForFeature(m.ctx, feature, m.providerID, m.market)
	if err != nil {
		m.err = err
		m.loading = false
		return
	}

	m.providerID = p.ID()
	m.market = actualMarket

	priceProvider, ok := p.(provider.PriceProvider)
	if !ok {
		m.err = fmt.Errorf("provider %s does not support price", p.ID())
		m.loading = false
		return
	}

	ccy := "usd"
	if m.market == model.MarketSpot || m.market == "" {
		if m.providerID == "binance" || m.providerID == "bitget" || m.providerID == "whitebit" {
			ccy = "usdt"
		}
	}

	res, err := priceProvider.Price(m.ctx, ids, ccy)
	if err != nil {
		m.err = err
		m.loading = false
		return
	}

	m.items = res.Data
	m.loading = false
}

func (m *PricesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "g":
			m.cursor = 0
		case "G":
			m.cursor = len(m.items) - 1
		}
	}

	return m, nil
}

func (m *PricesModel) View() string {
	if m.loading {
		return renderLoading("Fetching prices...", m.width, m.height)
	}

	if m.err != nil {
		return renderError(m.err.Error(), m.width, m.height)
	}

	if len(m.items) == 0 {
		return renderEmpty("No prices available", m.width, m.height)
	}

	providerLabel := m.providerID
	if providerLabel == "" {
		providerLabel = "auto"
	}

	header := fmt.Sprintf("Prices │ Provider: %s │ Market: %s",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Render(providerLabel),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Render(string(m.market)),
	)

	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n\n")

	headerLine := fmt.Sprintf("  %-6s %-20s %-15s %12s",
		"#", "Symbol", "Price", "24h Change")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#6B7280")).Render(headerLine))
	b.WriteString("\n")

	visibleRows := m.height - 8
	if visibleRows < 3 {
		visibleRows = 3
	}

	start := 0
	if m.cursor >= visibleRows {
		start = m.cursor - visibleRows + 1
	}
	end := start + visibleRows
	if end > len(m.items) {
		end = len(m.items)
	}

	for i := start; i < end; i++ {
		item := m.items[i]
		priceStr := formatPrice(item.Price)
		changeStr := formatChange(item.Change24h)

		row := fmt.Sprintf("  %-6d %-20s %-15s %s",
			i+1,
			truncate(item.Symbol, 20),
			priceStr,
			changeStr,
		)

		if i == m.cursor {
			b.WriteString(lipgloss.NewStyle().
				Background(lipgloss.Color("#3B82F6")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Render(" " + row))
		} else {
			b.WriteString(" " + row)
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render("j/k: navigate  │  g/G: first/last  │  q: quit"))

	return lipgloss.Place(m.width, m.height, lipgloss.Top, lipgloss.Left, b.String())
}

func (m *PricesModel) ProviderID() string       { return m.providerID }
func (m *PricesModel) Market() model.MarketType { return m.market }
func (m *PricesModel) Cursor() int              { return m.cursor }
func (m *PricesModel) Loading() bool            { return m.loading }
func (m *PricesModel) Error() error             { return m.err }

func (m *PricesModel) SetProvider(providerID string) {
	m.providerID = providerID
	m.loading = true
	m.items = nil
	go m.fetchPrices()
}

func (m *PricesModel) SetMarket(market model.MarketType) {
	m.market = market
	m.loading = true
	m.items = nil
	go m.fetchPrices()
}

func formatPrice(price float64) string {
	if price >= 1000 {
		return fmt.Sprintf("$%.2f", price)
	} else if price >= 1 {
		return fmt.Sprintf("$%.4f", price)
	}
	return fmt.Sprintf("$%.6f", price)
}

func formatChange(chg *float64) string {
	if chg == nil {
		return "—"
	}
	sign := "+"
	if *chg < 0 {
		sign = ""
	}
	color := "#10B981"
	if *chg < 0 {
		color = "#EF4444"
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(fmt.Sprintf("%s%.2f%%", sign, *chg))
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}

func renderLoading(msg string, width, height int) string {
	content := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED")).
		Render("bits — TUI\n\n")
	content += lipgloss.Place(width-4, height-4, lipgloss.Center, lipgloss.Center,
		lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render(msg))
	return content
}

func renderError(err string, width, height int) string {
	content := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED")).
		Render("bits — TUI\n\n")
	content += lipgloss.Place(width-4, height-4, lipgloss.Center, lipgloss.Center,
		lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Render("Error: "+err))
	return content
}

func renderEmpty(msg string, width, height int) string {
	content := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED")).
		Render("bits — TUI\n\n")
	content += lipgloss.Place(width-4, height-4, lipgloss.Center, lipgloss.Center,
		lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render(msg))
	return content
}
