package model

import "time"

type SymbolStatus string

const (
	SymbolStatusTrading SymbolStatus = "trading"
	SymbolStatusBreak   SymbolStatus = "break"
	SymbolStatusHalt    SymbolStatus = "halt"
)

type Symbol struct {
	Symbol         string
	BaseAsset      string // e.g. "BTC"
	QuoteAsset     string // e.g. "USDT"
	Status         SymbolStatus
	Market         MarketType
	PricePrecision *int
	QtyPrecision   *int
	MinPrice       *float64
	MaxPrice       *float64
	MinQty         *float64
	MaxQty         *float64
	StepSize       *float64 // quantity increment
	MakerFee       *float64
	TakerFee       *float64
	Extra          map[string]any
}

type ExchangeInfo struct {
	ExchangeID string
	Market     MarketType
	ServerTime *time.Time
	Symbols    []Symbol
	Extra      map[string]any
}

type ServerTime struct {
	Time      time.Time
	LocalTime *time.Time
	Latency   *time.Duration
	ClockSkew *time.Duration
	Extra     map[string]any
}
