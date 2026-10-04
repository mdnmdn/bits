package model

import "time"

// Trade is one public trade (a fill on the exchange tape).
type Trade struct {
	Symbol         string     `json:"sym"                yaml:"sym"                toon:"sym"`
	OriginalSymbol string     `json:"orig_sym,omitempty" yaml:"orig_sym,omitempty" toon:"orig_sym,omitempty"`
	Market         MarketType `json:"mkt"                yaml:"mkt"                toon:"mkt"`
	// ID is the provider's trade id (aggregate id on streams that aggregate).
	ID    string  `json:"id"  yaml:"id"  toon:"id"`
	Price float64 `json:"p"   yaml:"p"   toon:"p"`
	// Quantity is in base asset units.
	Quantity float64 `json:"q" yaml:"q" toon:"q"`
	// Side is the taker side: "buy" or "sell".
	Side string `json:"side" yaml:"side" toon:"side"`
	// Time is the trade time (UTC).
	Time time.Time `json:"t" yaml:"t" toon:"t"`
}

// Taker sides of a Trade.
const (
	TradeSideBuy  = "buy"
	TradeSideSell = "sell"
)
