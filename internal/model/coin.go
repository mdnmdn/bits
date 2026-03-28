package model

type CoinMarket struct {
	ID                string         `json:"id"               yaml:"id"               toon:"id"`
	Symbol            string         `json:"sym"              yaml:"sym"              toon:"sym"`
	Name              string         `json:"name"             yaml:"name"             toon:"name"`
	Currency          string         `json:"cur"              yaml:"cur"              toon:"cur"`
	Price             float64        `json:"price"            yaml:"price"            toon:"price"`
	MarketCap         *float64       `json:"mcap,omitempty"   yaml:"mcap,omitempty"   toon:"mcap,omitempty"`
	MarketCapRank     *int           `json:"rank,omitempty"   yaml:"rank,omitempty"   toon:"rank,omitempty"`
	Volume24h         *float64       `json:"vol24h,omitempty" yaml:"vol24h,omitempty" toon:"vol24h,omitempty"`
	PriceChangePct24h *float64       `json:"chg24h,omitempty" yaml:"chg24h,omitempty" toon:"chg24h,omitempty"`
	High24h           *float64       `json:"h24h,omitempty"   yaml:"h24h,omitempty"   toon:"h24h,omitempty"`
	Low24h            *float64       `json:"l24h,omitempty"   yaml:"l24h,omitempty"   toon:"l24h,omitempty"`
	Extra             map[string]any `json:"extra,omitempty"  yaml:"extra,omitempty"  toon:"extra,omitempty"`
}

type MarketOpts struct {
	Currency string
	PerPage  int
	Page     int
	Order    string
	Category string
}
