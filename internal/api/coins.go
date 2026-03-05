package api

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

func (c *Client) SimplePrice(ctx context.Context, ids []string, vsCurrency string) (PriceResponse, error) {
	params := url.Values{
		"ids":                 {strings.Join(ids, ",")},
		"vs_currencies":       {vsCurrency},
		"include_24hr_change": {"true"},
	}
	var result PriceResponse
	err := c.get(ctx, "/simple/price?"+params.Encode(), &result)
	return result, err
}

func (c *Client) CoinMarkets(ctx context.Context, vsCurrency string, perPage, page int, order, category string) ([]MarketCoin, error) {
	params := url.Values{
		"vs_currency": {vsCurrency},
		"per_page":    {fmt.Sprintf("%d", perPage)},
		"page":        {fmt.Sprintf("%d", page)},
		"order":       {order},
	}
	if category != "" {
		params.Set("category", category)
	}
	var result []MarketCoin
	err := c.get(ctx, "/coins/markets?"+params.Encode(), &result)
	return result, err
}

func (c *Client) Search(ctx context.Context, query string) (*SearchResponse, error) {
	params := url.Values{"query": {query}}
	var result SearchResponse
	err := c.get(ctx, "/search?"+params.Encode(), &result)
	return &result, err
}

func (c *Client) SearchTrending(ctx context.Context) (*TrendingResponse, error) {
	var result TrendingResponse
	err := c.get(ctx, "/search/trending", &result)
	return &result, err
}

func (c *Client) CoinHistory(ctx context.Context, id, date string) (*HistoricalData, error) {
	params := url.Values{"date": {date}, "localization": {"false"}}
	var result HistoricalData
	err := c.get(ctx, fmt.Sprintf("/coins/%s/history?%s", id, params.Encode()), &result)
	return &result, err
}

func (c *Client) CoinOHLC(ctx context.Context, id, vsCurrency string, days int) (OHLCData, error) {
	params := url.Values{
		"vs_currency": {vsCurrency},
		"days":        {fmt.Sprintf("%d", days)},
	}
	var result OHLCData
	err := c.get(ctx, fmt.Sprintf("/coins/%s/ohlc?%s", id, params.Encode()), &result)
	return result, err
}

func (c *Client) CoinMarketChartRange(ctx context.Context, id, vsCurrency string, from, to int64) (*MarketChartResponse, error) {
	params := url.Values{
		"vs_currency": {vsCurrency},
		"from":        {fmt.Sprintf("%d", from)},
		"to":          {fmt.Sprintf("%d", to)},
	}
	var result MarketChartResponse
	err := c.get(ctx, fmt.Sprintf("/coins/%s/market_chart/range?%s", id, params.Encode()), &result)
	return &result, err
}

func (c *Client) TopGainersLosers(ctx context.Context, vsCurrency, duration string, topCoins int) (*GainersLosersResponse, error) {
	if err := c.requirePaid(); err != nil {
		return nil, err
	}
	params := url.Values{
		"vs_currency": {vsCurrency},
		"duration":    {duration},
		"top_coins":   {fmt.Sprintf("%d", topCoins)},
	}
	var result GainersLosersResponse
	err := c.get(ctx, "/coins/top_gainers_losers?"+params.Encode(), &result)
	return &result, err
}

func (c *Client) CoinDetail(ctx context.Context, id string) (*CoinDetail, error) {
	params := url.Values{
		"localization":   {"false"},
		"tickers":        {"false"},
		"community_data": {"false"},
		"developer_data": {"false"},
	}
	var result CoinDetail
	err := c.get(ctx, fmt.Sprintf("/coins/%s?%s", id, params.Encode()), &result)
	return &result, err
}
