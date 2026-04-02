package mexc

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mdnmdn/bits/model"
)

// Price implements provider.PriceProvider.
func (c *Client) Price(ctx context.Context, symbols []string, currency string) (model.Response[[]model.CoinPrice], error) {
	resp := model.Response[[]model.CoinPrice]{
		Provider: providerID,
		Kind:     model.KindPrice,
	}

	if len(symbols) == 0 {
		resp.Market = model.MarketSpot
		return resp, nil
	}

	// Heuristic: if any requested symbol contains an underscore, it's likely a futures request.
	// This aligns with how bits routes internal requests.
	hasSpot := false
	hasFutures := false
	for _, s := range symbols {
		if strings.Contains(s, "_") {
			hasFutures = true
		} else {
			hasSpot = true
		}
	}

	if hasSpot && hasFutures {
		return resp, providerErr(model.ErrKindInvalidRequest, "MEXC provider does not support mixed spot/futures batches in a single Price() call", nil)
	}

	if hasFutures {
		resp.Market = model.MarketFutures
	} else {
		resp.Market = model.MarketSpot
	}

	var prices []model.CoinPrice
	for _, symbol := range symbols {
		market := model.MarketSpot
		if strings.Contains(symbol, "_") {
			market = model.MarketFutures
		}

		price, err := c.fetchPrice(ctx, symbol, market)
		if err != nil {
			resp.Errors = append(resp.Errors, model.ItemError{Symbol: symbol, Err: model.WrapError(providerID, err)})
			continue
		}
		prices = append(prices, price)
	}

	resp.Data = prices
	return resp, nil
}

func (c *Client) fetchPrice(ctx context.Context, symbol string, market model.MarketType) (model.CoinPrice, error) {
	if market == model.MarketFutures {
		data, err := c.doRequest(ctx, market, "/ticker", "symbol="+symbol)
		if err != nil {
			return model.CoinPrice{}, err
		}

		var tickerResp mexcFuturesTickerResponse
		if err := json.Unmarshal(data, &tickerResp); err != nil {
			return model.CoinPrice{}, providerErr(model.ErrKindParse, "failed to parse futures ticker response", err)
		}

		return model.CoinPrice{
			Symbol: tickerResp.Data.Symbol,
			Price:  tickerResp.Data.LastPrice,
		}, nil
	}

	// Spot / Margin
	data, err := c.doRequest(ctx, market, "/ticker/price", "symbol="+symbol)
	if err != nil {
		return model.CoinPrice{}, err
	}

	var ticker mexcSpotTicker
	if err := json.Unmarshal(data, &ticker); err != nil {
		return model.CoinPrice{}, providerErr(model.ErrKindParse, "failed to parse spot ticker response", err)
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
		data, err := c.doRequest(ctx, market, "/ticker", "symbol="+symbol)
		if err != nil {
			return resp, err
		}

		var tickerResp mexcFuturesTickerResponse
		if err := json.Unmarshal(data, &tickerResp); err != nil {
			return resp, providerErr(model.ErrKindParse, "failed to parse futures ticker response", err)
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
	data, err := c.doRequest(ctx, market, "/ticker/24hr", "symbol="+symbol)
	if err != nil {
		return resp, err
	}

	var t mexcSpotTicker24h
	if err := json.Unmarshal(data, &t); err != nil {
		return resp, providerErr(model.ErrKindParse, "failed to parse spot ticker response", err)
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

		data, err := c.doRequest(ctx, market, "/kline/"+symbol, query)
		if err != nil {
			return resp, err
		}

		var candlesResp mexcFuturesCandlesResponse
		if err := json.Unmarshal(data, &candlesResp); err != nil {
			return resp, providerErr(model.ErrKindParse, "failed to parse futures kline response", err)
		}

		var candles []model.Candle
		d := candlesResp.Data
		n := len(d.Time)
		// MEXC Futures return parallel slices; ensure they all have the same length
		// before indexing into them.
		if len(d.Open) < n || len(d.High) < n || len(d.Low) < n || len(d.Close) < n || len(d.Vol) < n {
			return resp, providerErr(model.ErrKindParse, "invalid futures kline response: inconsistent slice lengths", nil)
		}

		for i := 0; i < n; i++ {
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
	// MEXC API klines endpoint (/api/v3/klines) hasquirks that require workarounds:
	//
	// #1 Without any time filters:
	//    curl "https://api.mexc.com/api/v3/klines?symbol=SOLUSDT&interval=1d"
	//    -> Returns NEWEST 500 candles (correct)
	//
	// #2 With only startTime:
	//    curl "https://api.mexc.com/api/v3/klines?symbol=SOLUSDT&interval=1d&startTime=1774656000000"
	//    -> Returns OLDEST 500 candles (unexpected - should return newest from that time)
	//
	// #3 With only endTime:
	//    curl "https://api.mexc.com/api/v3/klines?symbol=SOLUSDT&interval=1d&endTime=1774828800000"
	//    -> Returns OLDEST candles up to that time (unexpected - should return newest)
	//
	// #4 With endTime + limit:
	//    curl "https://api.mexc.com/api/v3/klines?symbol=SOLUSDT&interval=1d&endTime=1774828800000&limit=10"
	//    -> Returns NEWEST 10 candles up to endTime (works correctly!)
	//
	// #5 With startTime + endTime:
	//    curl "https://api.mexc.com/api/v3/klines?symbol=SOLUSDT&interval=1d&startTime=1774656000000&endTime=1774828800000"
	//    -> Returns correct range (works correctly!)
	//
	// Workaround: Always fetch all (newest 500) and filter client-side.
	// This is the most reliable approach given the API quirks.
	mexcInterval := mapInterval(interval, false)
	query := fmt.Sprintf("symbol=%s&interval=%s", symbol, mexcInterval)

	data, err := c.doRequest(ctx, market, "/klines", query)
	if err != nil {
		return resp, err
	}

	var rawCandles [][]interface{}
	if err := json.Unmarshal(data, &rawCandles); err != nil {
		return resp, providerErr(model.ErrKindParse, "failed to parse spot kline response", err)
	}

	candles := parseSpotCandles(rawCandles)

	// Filter by time range if specified (MEXC API quirk workaround)
	if from := opts.From; from != nil && !from.IsZero() {
		filtered := make([]model.Candle, 0)
		for _, c := range candles {
			if !c.OpenTime.Before(*from) {
				filtered = append(filtered, c)
			}
		}
		candles = filtered
	}
	if to := opts.To; to != nil && !to.IsZero() {
		filtered := make([]model.Candle, 0)
		for _, c := range candles {
			if !c.OpenTime.After(*to) {
				filtered = append(filtered, c)
			}
		}
		candles = filtered
	}

	resp.Data = candles
	return resp, nil
}

func parseSpotCandles(raw [][]interface{}) []model.Candle {
	var candles []model.Candle
	for _, rc := range raw {
		if len(rc) < 6 {
			continue
		}

		// Safely parse time
		var t int64
		switch v := rc[0].(type) {
		case float64:
			t = int64(v)
		case string:
			t, _ = strconv.ParseInt(v, 10, 64)
		default:
			continue
		}

		// Helper for price fields
		parseFloat := func(val interface{}) float64 {
			switch v := val.(type) {
			case string:
				f, _ := strconv.ParseFloat(v, 64)
				return f
			case float64:
				return v
			default:
				return 0
			}
		}

		o := parseFloat(rc[1])
		h := parseFloat(rc[2])
		l := parseFloat(rc[3])
		cl := parseFloat(rc[4])
		v := parseFloat(rc[5])

		candles = append(candles, model.Candle{
			OpenTime: time.UnixMilli(t),
			Open:     o,
			High:     h,
			Low:      l,
			Close:    cl,
			Volume:   &v,
		})
	}
	return candles
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
		data, err := c.doRequest(ctx, market, "/depth/"+symbol, query)
		if err != nil {
			return resp, err
		}

		var obResp mexcFuturesOrderBookResponse
		if err := json.Unmarshal(data, &obResp); err != nil {
			return resp, providerErr(model.ErrKindParse, "failed to parse futures order book response", err)
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
	data, err := c.doRequest(ctx, market, "/depth", query)
	if err != nil {
		return resp, err
	}

	var ob mexcSpotOrderBook
	if err := json.Unmarshal(data, &ob); err != nil {
		return resp, providerErr(model.ErrKindParse, "failed to parse spot order book response", err)
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
	entries := make([]model.OrderBookEntry, 0, len(raw))
	for _, r := range raw {
		// MEXC Futures Depth documentation: [price, orders, quantity]
		// We need at least price and quantity (index 0 and 2).
		if len(r) < 3 {
			continue
		}
		entries = append(entries, model.OrderBookEntry{
			Price:    r[0],
			Quantity: r[2],
		})
	}
	return entries
}
