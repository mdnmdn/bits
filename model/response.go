package model

const (
	KindCandle       = "candle"
	KindTicker       = "ticker"
	KindPrice        = "price"
	KindOrderBook    = "orderbook"
	KindExchangeInfo = "exchange_info"
	KindServerTime   = "server_time"
	KindCoinMarket   = "coin_market"
)

// Response wraps any provider result with provenance metadata.
type Response[T any] struct {
	Kind              string         `json:"kind"                     yaml:"kind"                     toon:"kind"`
	Data              T              `json:"data"                    yaml:"data"                    toon:"data"`
	Provider          string         `json:"provider"                yaml:"provider"                toon:"provider"`               // provider that actually served this response
	Market            MarketType     `json:"mkt"                     yaml:"mkt"                     toon:"mkt"`                    // market that actually served this response
	Fallback          bool           `json:"fallback,omitempty"      yaml:"fallback,omitempty"      toon:"fallback,omitempty"`     // true when a different provider was selected automatically
	RequestedProvider string         `json:"req_provider,omitempty"  yaml:"req_provider,omitempty"  toon:"req_provider,omitempty"` // populated only when Fallback is true
	RequestedMarket   MarketType     `json:"req_mkt,omitempty"       yaml:"req_mkt,omitempty"       toon:"req_mkt,omitempty"`      // populated only when Fallback is true
	Metadata          map[string]any `json:"metadata,omitempty" yaml:"metadata,omitempty" toon:"metadata,omitempty"`
	Errors            []ItemError    `json:"errors,omitempty"       yaml:"errors,omitempty"        toon:"errors,omitempty"`
}

// ItemError pairs a symbol (or id) with the error that occurred for it.
type ItemError struct {
	Symbol string         `json:"sym" yaml:"sym" toon:"sym"`
	Err    *ProviderError `json:"err" yaml:"err" toon:"err"`
}
