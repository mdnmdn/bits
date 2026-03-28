package whitebit

import "github.com/mdnmdn/bits/internal/provider"

var _ provider.Provider = (*Client)(nil)
var _ provider.ExchangeProvider = (*Client)(nil)
var _ provider.PriceProvider = (*Client)(nil)
var _ provider.CandleProvider = (*Client)(nil)
var _ provider.TickerProvider = (*Client)(nil)
var _ provider.OrderBookProvider = (*Client)(nil)
var _ provider.PriceStreamProvider = (*Client)(nil)
var _ provider.OrderBookStreamProvider = (*Client)(nil)
