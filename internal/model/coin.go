package model

type CoinMarket struct {
	ID                string
	Symbol            string
	Name              string
	Currency          string
	Price             float64
	MarketCap         *float64
	MarketCapRank     *int
	Volume24h         *float64
	PriceChangePct24h *float64
	High24h           *float64
	Low24h            *float64
	Extra             map[string]any
}

type MarketOpts struct {
	Currency string
	PerPage  int
	Page     int
	Order    string
	Category string
}
