package provider

import (
	"context"

	"github.com/mdnmdn/bits/pkg/model"
)

// PriceStreamProvider streams live price updates.
type PriceStreamProvider interface {
	WatchPrices(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error)
}

// OrderBookStreamProvider streams live order book updates.
type OrderBookStreamProvider interface {
	WatchOrderBook(ctx context.Context, symbol string, market model.MarketType, depth int) (<-chan *model.OrderBook, error)
}
