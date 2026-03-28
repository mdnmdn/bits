package model

import "time"

type Candle struct {
	OpenTime  time.Time      `json:"t"              yaml:"t"              toon:"t"`
	Open      float64        `json:"o"              yaml:"o"              toon:"o"`
	High      float64        `json:"h"              yaml:"h"              toon:"h"`
	Low       float64        `json:"l"              yaml:"l"              toon:"l"`
	Close     float64        `json:"c"              yaml:"c"              toon:"c"`
	Volume    *float64       `json:"v,omitempty"    yaml:"v,omitempty"    toon:"v,omitempty"`    // base asset volume; absent in some aggregators
	CloseTime *time.Time     `json:"ct,omitempty"   yaml:"ct,omitempty"   toon:"ct,omitempty"`   // absent in some providers
	Extra     map[string]any `json:"extra,omitempty" yaml:"extra,omitempty" toon:"extra,omitempty"`
}

type CandleOpts struct {
	From  *time.Time
	To    *time.Time
	Limit *int
}
