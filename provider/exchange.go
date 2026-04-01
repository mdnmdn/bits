package provider

import (
	"context"

	"github.com/mdnmdn/bits/model"
)

// ExchangeProvider is implemented by direct exchange APIs (Binance, Bitget).
type ExchangeProvider interface {
	Provider
	ServerTime(ctx context.Context) (model.Response[model.ServerTime], error)
	ExchangeInfo(ctx context.Context, market model.MarketType) (model.Response[model.ExchangeInfo], error)
}
