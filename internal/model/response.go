package model

// Response wraps any provider result with provenance metadata.
type Response[T any] struct {
	Data              T
	Provider          string     // provider that actually served this response
	Market            MarketType // market that actually served this response
	Fallback          bool       // true when a different provider was selected automatically
	RequestedProvider string     // populated only when Fallback is true
	RequestedMarket   MarketType // populated only when Fallback is true
	Errors            []ItemError
}

// ItemError pairs a symbol (or id) with the error that occurred for it.
type ItemError struct {
	Symbol string
	Err    error
}
