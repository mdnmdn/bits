package coingecko

import "github.com/mdnmdn/bits/pkg/provider"

var _ provider.Provider = (*Client)(nil)
var _ provider.AggregatorProvider = (*Client)(nil)
var _ provider.PriceProvider = (*Client)(nil)
var _ provider.CandleProvider = (*Client)(nil)
var _ provider.PriceStreamProvider = (*Client)(nil)
