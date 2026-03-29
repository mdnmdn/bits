package model

import "time"

type Ticker24h struct {
	Symbol             string         `json:"sym"               yaml:"sym"               toon:"sym"`
	Market             MarketType     `json:"mkt"               yaml:"mkt"               toon:"mkt"`
	LastPrice          float64        `json:"last"              yaml:"last"              toon:"last"`
	PriceChange        *float64       `json:"chg,omitempty"     yaml:"chg,omitempty"     toon:"chg,omitempty"`
	PriceChangePercent *float64       `json:"chg_pct,omitempty" yaml:"chg_pct,omitempty" toon:"chg_pct,omitempty"`
	HighPrice          *float64       `json:"h,omitempty"       yaml:"h,omitempty"       toon:"h,omitempty"`
	LowPrice           *float64       `json:"l,omitempty"       yaml:"l,omitempty"       toon:"l,omitempty"`
	Volume             *float64       `json:"vol,omitempty"     yaml:"vol,omitempty"     toon:"vol,omitempty"` // base asset volume
	QuoteVolume        *float64       `json:"qvol,omitempty"    yaml:"qvol,omitempty"    toon:"qvol,omitempty"`
	OpenPrice          *float64       `json:"o,omitempty"       yaml:"o,omitempty"       toon:"o,omitempty"`
	WeightedAvgPrice   *float64       `json:"vwap,omitempty"    yaml:"vwap,omitempty"    toon:"vwap,omitempty"`
	BidPrice           *float64       `json:"bid,omitempty"     yaml:"bid,omitempty"     toon:"bid,omitempty"`
	AskPrice           *float64       `json:"ask,omitempty"     yaml:"ask,omitempty"     toon:"ask,omitempty"`
	OpenTime           *time.Time     `json:"ot,omitempty"      yaml:"ot,omitempty"      toon:"ot,omitempty"`
	CloseTime          *time.Time     `json:"ct,omitempty"      yaml:"ct,omitempty"      toon:"ct,omitempty"`
	Extra              map[string]any `json:"extra,omitempty"   yaml:"extra,omitempty"   toon:"extra,omitempty"`
}
