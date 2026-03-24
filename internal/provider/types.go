package provider

import (
	"context"

	"github.com/coingecko/coingecko-cli/internal/model"
)

// Provider is the base interface all data providers must implement.
type Provider interface {
	// Identity
	ID() string

	// Core Market Data
	SimplePrice(ctx context.Context, ids []string, vsCurrency string) (model.PriceResponse, error)
	CoinOHLC(ctx context.Context, id, vsCurrency, days, interval string) (model.OHLCData, error)

	// Configuration
	SetUserAgent(userAgent string)
}

// SymbolPricer supports price lookup by ticker symbol (e.g. "btc" -> bitcoin).
type SymbolPricer interface {
	SimplePriceBySymbols(ctx context.Context, symbols []string, vsCurrency string) (model.PriceResponse, error)
}

// MarketLister supports paginated market listings.
type MarketLister interface {
	CoinMarkets(ctx context.Context, vsCurrency string, perPage, page int, order, category string) ([]model.MarketCoin, error)
	FetchAllMarkets(ctx context.Context, vsCurrency string, total int, order, category string) ([]model.MarketCoin, error)
}

// Searcher supports coin search.
type Searcher interface {
	Search(ctx context.Context, query string) (*model.SearchResponse, error)
}

// TrendingProvider supports trending data.
type TrendingProvider interface {
	SearchTrending(ctx context.Context, showMax string) (*model.TrendingResponse, error)
}

// HistoricalProvider supports historical date-based and chart data.
type HistoricalProvider interface {
	CoinHistory(ctx context.Context, id, date string) (*model.HistoricalData, error)
	CoinMarketChart(ctx context.Context, id, vsCurrency, days, interval string) (*model.MarketChartResponse, error)
	CoinMarketChartRange(ctx context.Context, id, vsCurrency string, from, to int64, interval string) (*model.MarketChartResponse, error)
	CoinOHLCRange(ctx context.Context, id, vsCurrency string, from, to int64, interval string) (model.OHLCData, error)
}

// GainersLosersProvider supports top gainers/losers.
type GainersLosersProvider interface {
	TopGainersLosers(ctx context.Context, vsCurrency, duration, topCoins, priceChangePct string) (*model.GainersLosersResponse, error)
}

// DetailProvider supports detailed coin information.
type DetailProvider interface {
	CoinDetail(ctx context.Context, id string) (*model.CoinDetail, error)
}

// TickerProvider supports 24hr ticker statistics.
type TickerProvider interface {
	Ticker24h(ctx context.Context, symbol string) (*model.Ticker24h, error)
}

// OrderBookProvider supports order book depth data.
type OrderBookProvider interface {
	OrderBook(ctx context.Context, symbol string, limit int) (*model.OrderBook, error)
}
