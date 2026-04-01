package binance

import (
	_ "unsafe"

	"github.com/mdnmdn/bits/provider"
)

// Compile-time interface checks — no live API calls.
var _ provider.Provider = (*Client)(nil)
var _ provider.ExchangeProvider = (*Client)(nil)
var _ provider.PriceProvider = (*Client)(nil)
var _ provider.CandleProvider = (*Client)(nil)
var _ provider.TickerProvider = (*Client)(nil)
var _ provider.OrderBookProvider = (*Client)(nil)
var _ provider.OrderBookStreamProvider = (*Client)(nil)
