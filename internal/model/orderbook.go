package model

import "time"

type OrderBook struct {
	Symbol       string
	Market       MarketType
	Bids         []OrderBookEntry
	Asks         []OrderBookEntry
	LastUpdateID *int64
	Time         *time.Time
	Extra        map[string]any
}

type OrderBookEntry struct {
	Price    float64
	Quantity float64
}
