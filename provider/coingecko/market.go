package coingecko

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/mdnmdn/bits/model"
)

// Price fetches current prices for the given coin IDs.
func (c *Client) Price(ctx context.Context, ids []string, currency string) (model.Response[[]model.CoinPrice], error) {
	if currency == "" {
		currency = "usd"
	}

	params := url.Values{
		"ids":                 {strings.Join(ids, ",")},
		"vs_currencies":       {currency},
		"include_24hr_change": {"true"},
	}

	// Response: {"bitcoin":{"usd":50000,"usd_24h_change":1.5},...}
	var raw map[string]map[string]float64
	if err := c.get(ctx, "/simple/price?"+params.Encode(), &raw); err != nil {
		return model.Response[[]model.CoinPrice]{}, err
	}

	prices := make([]model.CoinPrice, 0, len(raw))
	for id, vals := range raw {
		price := vals[currency]
		changeKey := currency + "_24h_change"
		change := vals[changeKey]

		cp := model.CoinPrice{
			ID:        id,
			Symbol:    id,
			Currency:  currency,
			Price:     price,
			Change24h: &change,
		}
		prices = append(prices, cp)
	}

	return model.Response[[]model.CoinPrice]{
		Kind:     model.KindPrice,
		Data:     prices,
		Provider: providerID,
		Market:   model.MarketSpot,
	}, nil
}

// intervalToDays maps a candle interval string to a number of days for CoinGecko.
func intervalToDays(interval string, opts model.CandleOpts) string {
	if opts.Limit != nil && *opts.Limit > 0 {
		return fmt.Sprintf("%d", *opts.Limit)
	}
	switch interval {
	case "1h":
		return "2"
	case "4h":
		return "7"
	case "1d":
		return "30"
	case "1w":
		return "90"
	default:
		return "30"
	}
}

// Candles fetches OHLCV candle data for the given coin ID.
// CoinGecko returns [[ts_ms, open, high, low, close], ...].
func (c *Client) Candles(ctx context.Context, symbol string, market model.MarketType, interval string, opts model.CandleOpts) (model.Response[[]model.Candle], error) {
	days := intervalToDays(interval, opts)

	params := url.Values{
		"vs_currency": {"usd"},
		"days":        {days},
	}

	var raw [][]float64
	if err := c.get(ctx, fmt.Sprintf("/coins/%s/ohlc?%s", url.PathEscape(symbol), params.Encode()), &raw); err != nil {
		return model.Response[[]model.Candle]{}, err
	}

	candles := make([]model.Candle, 0, len(raw))
	for _, entry := range raw {
		if len(entry) < 5 {
			continue
		}
		ts := time.UnixMilli(int64(entry[0]))

		// Filter by From/To if provided.
		if opts.From != nil && ts.Before(*opts.From) {
			continue
		}
		if opts.To != nil && ts.After(*opts.To) {
			continue
		}

		candles = append(candles, model.Candle{
			OpenTime: ts,
			Open:     entry[1],
			High:     entry[2],
			Low:      entry[3],
			Close:    entry[4],
		})
	}

	return model.Response[[]model.Candle]{
		Kind:     model.KindCandle,
		Data:     candles,
		Provider: providerID,
		Market:   model.MarketSpot,
	}, nil
}
