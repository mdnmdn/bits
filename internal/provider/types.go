package provider

import (
	"context"

	"github.com/coingecko/coingecko-cli/internal/provider/coingecko"
)

// Provider defines the interface that all data providers must implement.
type Provider interface {
	// Identity
	ID() string

	// Market Data
	SimplePrice(ctx context.Context, ids []string, vsCurrency string) (coingecko.PriceResponse, error)
	SimplePriceBySymbols(ctx context.Context, symbols []string, vsCurrency string) (coingecko.PriceResponse, error)
	CoinMarkets(ctx context.Context, vsCurrency string, perPage, page int, order, category string) ([]coingecko.MarketCoin, error)
	FetchAllMarkets(ctx context.Context, vsCurrency string, total int, order, category string) ([]coingecko.MarketCoin, error)
	Search(ctx context.Context, query string) (*coingecko.SearchResponse, error)
	SearchTrending(ctx context.Context, showMax string) (*coingecko.TrendingResponse, error)
	CoinHistory(ctx context.Context, id, date string) (*coingecko.HistoricalData, error)
	CoinMarketChart(ctx context.Context, id, vsCurrency, days, interval string) (*coingecko.MarketChartResponse, error)
	CoinMarketChartRange(ctx context.Context, id, vsCurrency string, from, to int64, interval string) (*coingecko.MarketChartResponse, error)
	CoinOHLC(ctx context.Context, id, vsCurrency, days, interval string) (coingecko.OHLCData, error)
	CoinOHLCRange(ctx context.Context, id, vsCurrency string, from, to int64, interval string) (coingecko.OHLCData, error)
	TopGainersLosers(ctx context.Context, vsCurrency, duration, topCoins, priceChangePct string) (*coingecko.GainersLosersResponse, error)
	CoinDetail(ctx context.Context, id string) (*coingecko.CoinDetail, error)

	// Configuration
	SetUserAgent(userAgent string)
}
