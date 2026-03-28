package model

import "time"

type OrderBook struct {
	Symbol       string           `json:"sym"              yaml:"sym"              toon:"sym"`
	Market       MarketType       `json:"mkt"              yaml:"mkt"              toon:"mkt"`
	Bids         []OrderBookEntry `json:"bids"           yaml:"bids"             toon:"bids"`
	Asks         []OrderBookEntry `json:"asks"           yaml:"asks"             toon:"asks"`
	LastUpdateID *int64           `json:"uid,omitempty"    yaml:"uid,omitempty"    toon:"uid,omitempty"`
	Time         *time.Time       `json:"t,omitempty"      yaml:"t,omitempty"      toon:"t,omitempty"`
	Extra        map[string]any   `json:"extra,omitempty"  yaml:"extra,omitempty"  toon:"extra,omitempty"`
}

type OrderBookEntry struct {
	Price    float64 `json:"p"  yaml:"p"  toon:"p"`
	Quantity float64 `json:"q"  yaml:"q"  toon:"q"`
}
