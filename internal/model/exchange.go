package model

import "time"

type SymbolStatus string

func (s SymbolStatus) String() string { return string(s) }

const (
	SymbolStatusTrading SymbolStatus = "trading"
	SymbolStatusBreak   SymbolStatus = "break"
	SymbolStatusHalt    SymbolStatus = "halt"
)

type Symbol struct {
	Symbol         string         `json:"sym"              yaml:"sym"              toon:"sym"`
	BaseAsset      string         `json:"base"             yaml:"base"             toon:"base"`             // e.g. "BTC"
	QuoteAsset     string         `json:"quote"            yaml:"quote"            toon:"quote"`            // e.g. "USDT"
	Status         SymbolStatus   `json:"status"           yaml:"status"           toon:"status"`
	Market         MarketType     `json:"mkt"              yaml:"mkt"              toon:"mkt"`
	PricePrecision *int           `json:"pp,omitempty"     yaml:"pp,omitempty"     toon:"pp,omitempty"`
	QtyPrecision   *int           `json:"qp,omitempty"     yaml:"qp,omitempty"     toon:"qp,omitempty"`
	MinPrice       *float64       `json:"min_p,omitempty"  yaml:"min_p,omitempty"  toon:"min_p,omitempty"`
	MaxPrice       *float64       `json:"max_p,omitempty"  yaml:"max_p,omitempty"  toon:"max_p,omitempty"`
	MinQty         *float64       `json:"min_q,omitempty"  yaml:"min_q,omitempty"  toon:"min_q,omitempty"`
	MaxQty         *float64       `json:"max_q,omitempty"  yaml:"max_q,omitempty"  toon:"max_q,omitempty"`
	StepSize       *float64       `json:"step,omitempty"   yaml:"step,omitempty"   toon:"step,omitempty"`   // quantity increment
	MakerFee       *float64       `json:"maker,omitempty"  yaml:"maker,omitempty"  toon:"maker,omitempty"`
	TakerFee       *float64       `json:"taker,omitempty"  yaml:"taker,omitempty"  toon:"taker,omitempty"`
	Extra          map[string]any `json:"extra,omitempty"  yaml:"extra,omitempty"  toon:"extra,omitempty"`
}

type ExchangeInfo struct {
	ExchangeID string         `json:"exchange"         yaml:"exchange"         toon:"exchange"`
	Market     MarketType     `json:"mkt"              yaml:"mkt"              toon:"mkt"`
	ServerTime *time.Time     `json:"server_time,omitempty" yaml:"server_time,omitempty" toon:"server_time,omitempty"`
	Symbols    []Symbol       `json:"symbols"          yaml:"symbols"          toon:"symbols"`
	Extra      map[string]any `json:"extra,omitempty"  yaml:"extra,omitempty"  toon:"extra,omitempty"`
}

type ServerTime struct {
	Time      time.Time      `json:"t"               yaml:"t"               toon:"t"`
	LocalTime *time.Time     `json:"local,omitempty" yaml:"local,omitempty" toon:"local,omitempty"`
	Latency   *time.Duration `json:"lat,omitempty"   yaml:"lat,omitempty"   toon:"lat,omitempty"`
	ClockSkew *time.Duration `json:"skew,omitempty"  yaml:"skew,omitempty"  toon:"skew,omitempty"`
	Extra     map[string]any `json:"extra,omitempty" yaml:"extra,omitempty" toon:"extra,omitempty"`
}
