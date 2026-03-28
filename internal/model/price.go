package model

type CoinPrice struct {
	ID        string         `json:"id"              yaml:"id"              toon:"id"` // coin id (aggregators) or trading symbol (exchanges)
	Symbol    string         `json:"sym"             yaml:"sym"             toon:"sym"`
	Currency  string         `json:"cur"             yaml:"cur"             toon:"cur"`
	Price     float64        `json:"price"           yaml:"price"           toon:"price"`
	Change24h *float64       `json:"chg24h,omitempty" yaml:"chg24h,omitempty" toon:"chg24h,omitempty"` // percent; optional
	Extra     map[string]any `json:"extra,omitempty" yaml:"extra,omitempty" toon:"extra,omitempty"`
}
