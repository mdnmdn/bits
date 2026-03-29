package provider

import (
	"context"

	"github.com/mdnmdn/bits/pkg/model"
)

// PriceProvider fetches current prices from a crypto data source.
// Aggregators (like CoinGecko) typically use coin IDs, while exchanges use trading symbols.
// This interface is batch-native, meaning it can receive multiple IDs/symbols at once.
type PriceProvider interface {
	// Price retrieves the current price for the given IDs or symbols.
	// The currency parameter is primarily used by aggregators.
	Price(ctx context.Context, ids []string, currency string) (model.Response[[]model.CoinPrice], error)
}

// CandleProvider fetches OHLCV candle data.
type CandleProvider interface {
	Candles(ctx context.Context, symbol string, market model.MarketType, interval string, opts model.CandleOpts) (model.Response[[]model.Candle], error)
}

// TickerProvider fetches 24h rolling ticker statistics.
// Single-symbol: the resolver fans out for multi-symbol calls.
type TickerProvider interface {
	Ticker24h(ctx context.Context, symbol string, market model.MarketType) (model.Response[model.Ticker24h], error)
}

// OrderBookProvider fetches order book depth snapshots.
// Single-symbol: the resolver fans out for multi-symbol calls.
type OrderBookProvider interface {
	OrderBook(ctx context.Context, symbol string, market model.MarketType, depth int) (model.Response[model.OrderBook], error)
}
