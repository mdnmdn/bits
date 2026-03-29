package tui

import (
	"time"

	"github.com/mdnmdn/bits/pkg/model"
)

type Section string

const (
	SectionPrices   Section = "prices"
	SectionTicker   Section = "ticker"
	SectionExchange Section = "exchange"
	SectionBook     Section = "book"
	SectionCandles  Section = "candles"
	SectionMarkets  Section = "markets"
	SectionTrending Section = "trending"
)

func (s Section) Feature() string {
	switch s {
	case SectionPrices:
		return "price"
	case SectionTicker:
		return "ticker_24h"
	case SectionExchange:
		return "server_time"
	case SectionBook:
		return "order_book"
	case SectionCandles:
		return "candles"
	case SectionMarkets:
		return "markets_list"
	case SectionTrending:
		return "trending"
	default:
		return ""
	}
}

type RefreshConfig struct {
	Interval   time.Duration
	Enabled    bool
	LastUpdate time.Time
	TickCmd    chan time.Time
}

func ParseRefreshInterval(s string) time.Duration {
	switch s {
	case "5s":
		return 5 * time.Second
	case "10s":
		return 10 * time.Second
	case "30s":
		return 30 * time.Second
	case "1m":
		return 1 * time.Minute
	default:
		return 0
	}
}

type FilterState struct {
	Symbol     string
	ShowFilter bool
}

type TUIState struct {
	Section   Section
	Provider  string
	Market    model.MarketType
	Filters   FilterState
	Providers []string
	Markets   []model.MarketType
	MultiMode bool
	Refresh   RefreshConfig
	Error     error
}

func NewTUIState(opts Options) *TUIState {
	section := Section(opts.Section)
	if section == "" {
		section = SectionPrices
	}

	state := &TUIState{
		Section:  section,
		Provider: opts.Provider,
		Market:   model.MarketType(opts.Market),
		Filters: FilterState{
			Symbol:     opts.Symbol,
			ShowFilter: false,
		},
		MultiMode: false,
	}

	if opts.RefreshInterval != "" {
		state.Refresh.Interval = ParseRefreshInterval(opts.RefreshInterval)
		state.Refresh.Enabled = state.Refresh.Interval > 0
	}

	return state
}
