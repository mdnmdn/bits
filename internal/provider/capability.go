package provider

import (
	"context"

	"github.com/mdnmdn/bits/internal/model"
)

// PriceProvider fetches current prices. Aggregators use coin IDs; exchanges use symbols.
// Batch-native: provider receives all ids at once.
type PriceProvider interface {
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
