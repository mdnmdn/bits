package mexc

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mdnmdn/bits/pkg/model"
)

// Price implements provider.PriceProvider.
func (c *Client) Price(ctx context.Context, symbols []string, currency string) (model.Response[[]model.CoinPrice], error) {
	resp := model.Response[[]model.CoinPrice]{
		Provider: providerID,
		Market:   model.MarketSpot,
		Kind:     model.KindPrice,
	}

	if len(symbols) == 0 {
		return resp, nil
	}

	var prices []model.CoinPrice
	for _, symbol := range symbols {
		market := model.MarketSpot
		if strings.Contains(symbol, "_") {
			market = model.MarketFutures
		}

		price, err := c.fetchPrice(ctx, symbol, market)
		if err != nil {
			resp.Errors = append(resp.Errors, model.ItemError{Symbol: symbol, Err: err})
			continue
		}
		prices = append(prices, price)
	}

	resp.Data = prices
	return resp, nil
}

func (c *Client) fetchPrice(ctx context.Context, symbol string, market model.MarketType) (model.CoinPrice, error) {
	if market == model.MarketFutures {
		data, err := c.doRequest(market, "/ticker", "symbol="+symbol)
		if err != nil {
			return model.CoinPrice{}, err
		}

		var tickerResp mexcFuturesTickerResponse
		if err := json.Unmarshal(data, &tickerResp); err != nil {
			return model.CoinPrice{}, err
		}

		return model.CoinPrice{
			Symbol: tickerResp.Data.Symbol,
			Price:  tickerResp.Data.LastPrice,
		}, nil
	}

	// Spot / Margin
	data, err := c.doRequest(market, "/ticker/price", "symbol="+symbol)
	if err != nil {
		return model.CoinPrice{}, err
	}

	var ticker mexcSpotTicker
	if err := json.Unmarshal(data, &ticker); err != nil {
		return model.CoinPrice{}, err
	}

	price, _ := strconv.ParseFloat(ticker.Price, 64)
	return model.CoinPrice{
		Symbol: ticker.Symbol,
		Price:  price,
	}, nil
}

// Ticker24h implements provider.TickerProvider.
func (c *Client) Ticker24h(ctx context.Context, symbol string, market model.MarketType) (model.Response[model.Ticker24h], error) {
	resp := model.Response[model.Ticker24h]{
		Provider: providerID,
		Market:   market,
		Kind:     model.KindTicker,
	}

	if market == model.MarketFutures {
		data, err := c.doRequest(market, "/ticker", "symbol="+symbol)
		if err != nil {
			return resp, err
		}

		var tickerResp mexcFuturesTickerResponse
		if err := json.Unmarshal(data, &tickerResp); err != nil {
			return resp, err
		}

		t := tickerResp.Data
		resp.Data = model.Ticker24h{
			Symbol:             t.Symbol,
			Market:             market,
			LastPrice:          t.LastPrice,
			HighPrice:          &t.High24Price,
			LowPrice:           &t.Lower24Price,
			Volume:             &t.Volume24,
			QuoteVolume:        &t.Amount24,
			PriceChange:        &t.RiseFallValue,
			PriceChangePercent: &t.RiseFallRate,
		}
		return resp, nil
	}

	// Spot / Margin
	data, err := c.doRequest(market, "/ticker/24hr", "symbol="+symbol)
	if err != nil {
		return resp, err
	}

	var t mexcSpotTicker24h
	if err := json.Unmarshal(data, &t); err != nil {
		return resp, err
	}

	last, _ := strconv.ParseFloat(t.LastPrice, 64)
	high, _ := strconv.ParseFloat(t.HighPrice, 64)
	low, _ := strconv.ParseFloat(t.LowPrice, 64)
	vol, _ := strconv.ParseFloat(t.Volume, 64)
	qVol, _ := strconv.ParseFloat(t.QuoteVolume, 64)
	change, _ := strconv.ParseFloat(t.PriceChange, 64)
	percent, _ := strconv.ParseFloat(t.PriceChangePercent, 64)

	resp.Data = model.Ticker24h{
		Symbol:             t.Symbol,
		Market:             market,
		LastPrice:          last,
		HighPrice:          &high,
		LowPrice:           &low,
		Volume:             &vol,
		QuoteVolume:        &qVol,
		PriceChange:        &change,
		PriceChangePercent: &percent,
	}

	return resp, nil
}

// Candles implements provider.CandleProvider.
func (c *Client) Candles(ctx context.Context, symbol string, market model.MarketType, interval string, opts model.CandleOpts) (model.Response[[]model.Candle], error) {
	resp := model.Response[[]model.Candle]{
		Provider: providerID,
		Market:   market,
		Kind:     model.KindCandle,
	}

	if market == model.MarketFutures {
		mexcInterval := mapInterval(interval, true)
		query := fmt.Sprintf("interval=%s", mexcInterval)
		if opts.Limit != nil {
			query += fmt.Sprintf("&limit=%d", *opts.Limit)
		}
		if opts.From != nil && !opts.From.IsZero() {
			query += fmt.Sprintf("&start=%d", opts.From.Unix())
		}
		if opts.To != nil && !opts.To.IsZero() {
			query += fmt.Sprintf("&end=%d", opts.To.Unix())
		}

		data, err := c.doRequest(market, "/kline/"+symbol, query)
		if err != nil {
			return resp, err
		}

		var candlesResp mexcFuturesCandlesResponse
		if err := json.Unmarshal(data, &candlesResp); err != nil {
			return resp, err
		}

		var candles []model.Candle
		d := candlesResp.Data
		for i := range d.Time {
			v := d.Vol[i]
			candles = append(candles, model.Candle{
				OpenTime: time.Unix(d.Time[i], 0),
				Open:     d.Open[i],
				High:     d.High[i],
				Low:      d.Low[i],
				Close:    d.Close[i],
				Volume:   &v,
			})
		}
		resp.Data = candles
		return resp, nil
	}

	// Spot / Margin
	mexcInterval := mapInterval(interval, false)
	query := fmt.Sprintf("symbol=%s&interval=%s", symbol, mexcInterval)
	if opts.Limit != nil {
		query += fmt.Sprintf("&limit=%d", *opts.Limit)
	}
	if opts.From != nil && !opts.From.IsZero() {
		query += fmt.Sprintf("&startTime=%d", opts.From.UnixMilli())
	}
	if opts.To != nil && !opts.To.IsZero() {
		query += fmt.Sprintf("&endTime=%d", opts.To.UnixMilli())
	}

	data, err := c.doRequest(market, "/klines", query)
	if err != nil {
		return resp, err
	}

	var rawCandles [][]interface{}
	if err := json.Unmarshal(data, &rawCandles); err != nil {
		return resp, err
	}

	var candles []model.Candle
	for _, rc := range rawCandles {
		if len(rc) < 6 {
			continue
		}
		t := int64(rc[0].(float64))
		o, _ := strconv.ParseFloat(rc[1].(string), 64)
		h, _ := strconv.ParseFloat(rc[2].(string), 64)
		l, _ := strconv.ParseFloat(rc[3].(string), 64)
		cl, _ := strconv.ParseFloat(rc[4].(string), 64)
		v, _ := strconv.ParseFloat(rc[5].(string), 64)

		candles = append(candles, model.Candle{
			OpenTime: time.UnixMilli(t),
			Open:     o,
			High:     h,
			Low:      l,
			Close:    cl,
			Volume:   &v,
		})
	}

	resp.Data = candles
	return resp, nil
}

// OrderBook implements provider.OrderBookProvider.
func (c *Client) OrderBook(ctx context.Context, symbol string, market model.MarketType, depth int) (model.Response[model.OrderBook], error) {
	resp := model.Response[model.OrderBook]{
		Provider: providerID,
		Market:   market,
		Kind:     model.KindOrderBook,
	}

	if market == model.MarketFutures {
		query := ""
		if depth > 0 {
			query = fmt.Sprintf("limit=%d", depth)
		}
		data, err := c.doRequest(market, "/depth/"+symbol, query)
		if err != nil {
			return resp, err
		}

		var obResp mexcFuturesOrderBookResponse
		if err := json.Unmarshal(data, &obResp); err != nil {
			return resp, err
		}

		resp.Data = model.OrderBook{
			Symbol: symbol,
			Bids:   parseFuturesOrders(obResp.Data.Bids),
			Asks:   parseFuturesOrders(obResp.Data.Asks),
		}
		return resp, nil
	}

	// Spot / Margin
	query := "symbol=" + symbol
	if depth > 0 {
		query += fmt.Sprintf("&limit=%d", depth)
	}
	data, err := c.doRequest(market, "/depth", query)
	if err != nil {
		return resp, err
	}

	var ob mexcSpotOrderBook
	if err := json.Unmarshal(data, &ob); err != nil {
		return resp, err
	}

	resp.Data = model.OrderBook{
		Symbol: symbol,
		Bids:   parseSpotOrders(ob.Bids),
		Asks:   parseSpotOrders(ob.Asks),
	}

	return resp, nil
}

func mapInterval(interval string, futures bool) string {
	if futures {
		switch interval {
		case "1m":
			return "Min1"
		case "5m":
			return "Min5"
		case "15m":
			return "Min15"
		case "30m":
			return "Min30"
		case "1h":
			return "Min60"
		case "4h":
			return "Hour4"
		case "8h":
			return "Hour8"
		case "1d":
			return "Day1"
		case "1w":
			return "Week1"
		case "1M":
			return "Month1"
		default:
			return "Min60"
		}
	}
	// Spot intervals
	switch interval {
	case "1m":
		return "1m"
	case "5m":
		return "5m"
	case "15m":
		return "15m"
	case "30m":
		return "30m"
	case "1h":
		return "60m"
	case "4h":
		return "4h"
	case "1d":
		return "1d"
	case "1w":
		return "1w"
	case "1M":
		return "1M"
	default:
		return "60m"
	}
}

func parseSpotOrders(raw [][]string) []model.OrderBookEntry {
	entries := make([]model.OrderBookEntry, len(raw))
	for i, r := range raw {
		price, _ := strconv.ParseFloat(r[0], 64)
		amount, _ := strconv.ParseFloat(r[1], 64)
		entries[i] = model.OrderBookEntry{Price: price, Quantity: amount}
	}
	return entries
}

func parseFuturesOrders(raw [][]float64) []model.OrderBookEntry {
	entries := make([]model.OrderBookEntry, len(raw))
	for i, r := range raw {
		if len(r) >= 3 {
			entries[i] = model.OrderBookEntry{Price: r[0], Quantity: r[2]}
		} else if len(r) >= 1 {
			entries[i] = model.OrderBookEntry{Price: r[0]}
		}
	}
	return entries
}
