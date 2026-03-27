package model

import "time"

type Ticker24h struct {
	Symbol             string
	Market             MarketType
	LastPrice          float64
	PriceChange        *float64
	PriceChangePercent *float64
	HighPrice          *float64
	LowPrice           *float64
	Volume             *float64 // base asset volume
	QuoteVolume        *float64
	OpenPrice          *float64
	WeightedAvgPrice   *float64
	BidPrice           *float64
	AskPrice           *float64
	OpenTime           *time.Time
	CloseTime          *time.Time
	Extra              map[string]any
}
