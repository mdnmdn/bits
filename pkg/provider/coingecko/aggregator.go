package coingecko

import (
	"context"
	"fmt"
	"net/url"

	"github.com/mdnmdn/bits/pkg/model"
)

// apiMarketCoin is the JSON shape returned by /coins/markets.
type apiMarketCoin struct {
	ID                       string  `json:"id"`
	Symbol                   string  `json:"symbol"`
	Name                     string  `json:"name"`
	CurrentPrice             float64 `json:"current_price"`
	MarketCap                float64 `json:"market_cap"`
	MarketCapRank            int     `json:"market_cap_rank"`
	TotalVolume              float64 `json:"total_volume"`
	PriceChangePercentage24h float64 `json:"price_change_percentage_24h"`
	High24h                  float64 `json:"high_24h"`
	Low24h                   float64 `json:"low_24h"`
}

// CoinMarkets fetches a paginated list of coins with market data.
func (c *Client) CoinMarkets(ctx context.Context, opts model.MarketOpts) (model.Response[[]model.CoinMarket], error) {
	// Apply defaults.
	currency := opts.Currency
	if currency == "" {
		currency = "usd"
	}
	perPage := opts.PerPage
	if perPage <= 0 {
		perPage = 100
	}
	page := opts.Page
	if page <= 0 {
		page = 1
	}
	order := opts.Order
	if order == "" {
		order = "market_cap_desc"
	}

	params := url.Values{
		"vs_currency": {currency},
		"per_page":    {fmt.Sprintf("%d", perPage)},
		"page":        {fmt.Sprintf("%d", page)},
		"order":       {order},
	}
	if opts.Category != "" {
		params.Set("category", opts.Category)
	}

	var raw []apiMarketCoin
	if err := c.get(ctx, "/coins/markets?"+params.Encode(), &raw); err != nil {
		return model.Response[[]model.CoinMarket]{}, err
	}

	coins := make([]model.CoinMarket, len(raw))
	for i, r := range raw {
		mc := r.MarketCap
		rank := r.MarketCapRank
		vol := r.TotalVolume
		pct := r.PriceChangePercentage24h
		high := r.High24h
		low := r.Low24h

		coins[i] = model.CoinMarket{
			ID:                r.ID,
			Symbol:            r.Symbol,
			Name:              r.Name,
			Currency:          currency,
			Price:             r.CurrentPrice,
			MarketCap:         &mc,
			MarketCapRank:     &rank,
			Volume24h:         &vol,
			PriceChangePct24h: &pct,
			High24h:           &high,
			Low24h:            &low,
		}
	}

	return model.Response[[]model.CoinMarket]{
		Kind:     model.KindCoinMarket,
		Data:     coins,
		Provider: providerID,
		Market:   model.MarketSpot,
	}, nil
}
