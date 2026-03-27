package provider

import (
	"context"

	"github.com/mdnmdn/bits/internal/model"
)

// AggregatorProvider is implemented by market data aggregators (CoinGecko, CoinMarketCap).
type AggregatorProvider interface {
	Provider
	CoinMarkets(ctx context.Context, opts model.MarketOpts) (model.Response[[]model.CoinMarket], error)
}
