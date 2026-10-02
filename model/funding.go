package model

import "time"

// FundingRate is a single perpetual-futures funding settlement.
type FundingRate struct {
	Symbol string `json:"sym" yaml:"sym" toon:"sym"`
	// Time is the funding settlement time.
	Time time.Time `json:"t" yaml:"t" toon:"t"`
	// Rate is the funding rate for the period as a ratio (0.0001 = 0.01%).
	Rate float64 `json:"rate" yaml:"rate" toon:"rate"`
	// MarkPrice is the mark price at settlement; absent in some providers.
	MarkPrice *float64 `json:"mark,omitempty" yaml:"mark,omitempty" toon:"mark,omitempty"`
}

// FundingRateOpts bounds a funding-rate history query.
// Results are returned in ascending time order. With From set and Limit
// smaller than the window, the earliest Limit entries at or after From are
// returned; without From, the most recent Limit entries.
type FundingRateOpts struct {
	From  *time.Time
	To    *time.Time
	Limit *int
}
