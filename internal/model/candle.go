package model

import "time"

type Candle struct {
	OpenTime  time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    *float64   // base asset volume; absent in some aggregators
	CloseTime *time.Time // absent in some providers
	Extra     map[string]any
}

type CandleOpts struct {
	From  *time.Time
	To    *time.Time
	Limit *int
}
