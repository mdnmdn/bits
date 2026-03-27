package model

type CoinPrice struct {
	ID        string // coin id (aggregators) or trading symbol (exchanges)
	Symbol    string
	Currency  string
	Price     float64
	Change24h *float64 // percent; optional
	Extra     map[string]any
}
