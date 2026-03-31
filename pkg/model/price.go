package model

import "time"

// CoinPrice represents the current price of a cryptocurrency.
// It includes information about the coin's ID, symbol, current price, and 24h change.
type CoinPrice struct {
	// ID is the unique identifier for the coin (e.g., "bitcoin") or trading symbol (e.g., "BTCUSDT").
	ID        string   `json:"id"              yaml:"id"              toon:"id"` // coin id (aggregators) or trading symbol (exchanges)
	Symbol    string   `json:"sym"             yaml:"sym"             toon:"sym"`
	Currency  string   `json:"cur"             yaml:"cur"             toon:"cur"`
	Price     float64  `json:"price"           yaml:"price"           toon:"price"`
	Change24h *float64 `json:"chg24h,omitempty" yaml:"chg24h,omitempty" toon:"chg24h,omitempty"` // percent; optional

	Volume24h *float64 `json:"vol24h,omitempty" yaml:"vol24h,omitempty" toon:"vol24h,omitempty"`
	High24h   *float64 `json:"high24h,omitempty" yaml:"high24h,omitempty" toon:"high24h,omitempty"`
	Low24h    *float64 `json:"low24h,omitempty"  yaml:"low24h,omitempty"  toon:"low24h,omitempty"`
	Open24h   *float64 `json:"open24h,omitempty" yaml:"open24h,omitempty" toon:"open24h,omitempty"`

	BidPrice *float64 `json:"bidPr,omitempty" yaml:"bidPr,omitempty" toon:"bidPr,omitempty"`
	BidSize  *float64 `json:"bidSz,omitempty" yaml:"bidSz,omitempty" toon:"bidSz,omitempty"`
	AskPrice *float64 `json:"askPr,omitempty" yaml:"askPr,omitempty" toon:"askPr,omitempty"`
	AskSize  *float64 `json:"askSz,omitempty" yaml:"askSz,omitempty" toon:"askSz,omitempty"`

	Time  *time.Time     `json:"t,omitempty" yaml:"t,omitempty" toon:"t,omitempty"`
	Extra map[string]any `json:"extra,omitempty" yaml:"extra,omitempty" toon:"extra,omitempty"`
}
