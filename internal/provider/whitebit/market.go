package whitebit

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mdnmdn/bits/internal/model"
)

// whitebitTicker represents a single market ticker from the v1 public API.
// Endpoint: GET /api/v1/public/ticker?market={symbol}
type whitebitTicker struct {
	Open   string `json:"open"`
	High   string `json:"high"`
	Low    string `json:"low"`
	Last   string `json:"last"`
	Volume string `json:"volume"`
	Deal   string `json:"deal"`
	Bid    string `json:"bid"`
	Ask    string `json:"ask"`
	Change string `json:"change"`
}

// whitebitV1TickerResponse wraps the v1 API envelope.
type whitebitV1TickerResponse struct {
	Success bool           `json:"success"`
	Result  whitebitTicker `json:"result"`
}

// whitebitCandleResponse is the response envelope for kline (candle) data.
type whitebitCandleResponse struct {
	Success bool    `json:"success"`
	Result  [][]any `json:"result"`
}

// whitebitOrderBookResponse is the response envelope for order book data.
// Entries are [price, qty] string pairs; timestamp is Unix seconds.
type whitebitOrderBookResponse struct {
	Timestamp int64      `json:"timestamp"`
	Asks      [][]string `json:"asks"`
	Bids      [][]string `json:"bids"`
}

// fetchTicker fetches a single market ticker via the v1 API which includes open/high/low.
func (c *Client) fetchTicker(symbol string) (*whitebitTicker, error) {
	body, err := c.doRequest("/api/v1/public/ticker?market=" + symbol)
	if err != nil {
		return nil, err
	}
	var resp whitebitV1TickerResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse ticker response: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("ticker not found for symbol %s", symbol)
	}
	return &resp.Result, nil
}

// translateFuturesSymbol converts a spot symbol to a futures symbol (e.g. BTC_USDT -> BTC_PERP)
func translateFuturesSymbol(symbol string) string {
	if strings.HasSuffix(symbol, "_USDT") {
		return strings.TrimSuffix(symbol, "_USDT") + "_PERP"
	}
	return symbol
}

// Price fetches current prices for the given symbols.
// ids are trading symbols (e.g. "BTC_USDT"). currency is used as metadata only.
func (c *Client) Price(ctx context.Context, ids []string, currency string) (model.Response[[]model.CoinPrice], error) {
	prices := make([]model.CoinPrice, 0, len(ids))
	var itemErrors []model.ItemError

	for _, symbol := range ids {
		ticker, err := c.fetchTicker(symbol)
		if err != nil {
			itemErrors = append(itemErrors, model.ItemError{Symbol: symbol, Err: err})
			continue
		}
		price, _ := strconv.ParseFloat(ticker.Last, 64)
		changePct, _ := strconv.ParseFloat(ticker.Change, 64)
		prices = append(prices, model.CoinPrice{
			ID:        symbol,
			Symbol:    symbol,
			Currency:  currency,
			Price:     price,
			Change24h: &changePct,
		})
	}

	return model.Response[[]model.CoinPrice]{
		Kind:     model.KindPrice,
		Provider: providerID,
		Market:   model.MarketSpot, // Price usually defaults to spot in this app's context if not specified
		Data:     prices,
		Errors:   itemErrors,
	}, nil
}

// Ticker24h fetches 24-hour rolling ticker statistics for the given symbol.
func (c *Client) Ticker24h(ctx context.Context, symbol string, market model.MarketType) (model.Response[model.Ticker24h], error) {
	if market == model.MarketFutures {
		return c.futuresTicker24h(ctx, symbol, market)
	}

	ticker, err := c.fetchTicker(symbol)
	if err != nil {
		return model.Response[model.Ticker24h]{}, err
	}

	lastPrice, _ := strconv.ParseFloat(ticker.Last, 64)
	highPrice, _ := strconv.ParseFloat(ticker.High, 64)
	lowPrice, _ := strconv.ParseFloat(ticker.Low, 64)
	openPrice, _ := strconv.ParseFloat(ticker.Open, 64)
	baseVol, _ := strconv.ParseFloat(ticker.Volume, 64)
	quoteVol, _ := strconv.ParseFloat(ticker.Deal, 64)
	changePct, _ := strconv.ParseFloat(ticker.Change, 64)
	bidPrice, _ := strconv.ParseFloat(ticker.Bid, 64)
	askPrice, _ := strconv.ParseFloat(ticker.Ask, 64)
	priceChange := lastPrice - openPrice

	t24h := model.Ticker24h{
		Symbol:             symbol,
		Market:             market,
		LastPrice:          lastPrice,
		PriceChange:        &priceChange,
		PriceChangePercent: &changePct,
		HighPrice:          &highPrice,
		LowPrice:           &lowPrice,
		OpenPrice:          &openPrice,
		Volume:             &baseVol,
		QuoteVolume:        &quoteVol,
		BidPrice:           &bidPrice,
		AskPrice:           &askPrice,
	}

	return model.Response[model.Ticker24h]{
		Kind:     model.KindTicker,
		Provider: providerID,
		Market:   market,
		Data:     t24h,
	}, nil
}

// Candles fetches OHLCV candle data for the given symbol and interval.
// Column order in WhiteBit response: [ts, open, close, high, low, vol, amount]
func (c *Client) Candles(_ context.Context, symbol string, market model.MarketType, interval string, opts model.CandleOpts) (model.Response[[]model.Candle], error) {
	query := fmt.Sprintf("market=%s&interval=%s", symbol, convertInterval(interval))

	if opts.Limit != nil {
		query += fmt.Sprintf("&limit=%d", *opts.Limit)
	}
	if opts.From != nil {
		query += fmt.Sprintf("&start=%d", opts.From.Unix())
	}
	if opts.To != nil {
		query += fmt.Sprintf("&end=%d", opts.To.Unix())
	}

	body, err := c.doRequest("/api/v1/public/kline?" + query)
	if err != nil {
		return model.Response[[]model.Candle]{}, err
	}

	var resp whitebitCandleResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return model.Response[[]model.Candle]{}, fmt.Errorf("failed to parse candles response: %w", err)
	}

	if !resp.Success {
		return model.Response[[]model.Candle]{}, fmt.Errorf("API returned success=false")
	}

	candles := make([]model.Candle, 0, len(resp.Result))
	for i, row := range resp.Result {
		if len(row) < 5 {
			return model.Response[[]model.Candle]{}, fmt.Errorf("invalid candle data at index %d: expected at least 5 fields, got %d", i, len(row))
		}

		// Column order: [ts, open, close, high, low, vol, amount]
		ts := int64(row[0].(float64))
		openTime := time.Unix(ts, 0)
		open, _ := strconv.ParseFloat(fmt.Sprint(row[1]), 64)
		close, _ := strconv.ParseFloat(fmt.Sprint(row[2]), 64)
		high, _ := strconv.ParseFloat(fmt.Sprint(row[3]), 64)
		low, _ := strconv.ParseFloat(fmt.Sprint(row[4]), 64)

		var vol *float64
		if len(row) >= 6 {
			v, _ := strconv.ParseFloat(fmt.Sprint(row[5]), 64)
			vol = &v
		}

		candles = append(candles, model.Candle{
			OpenTime: openTime,
			Open:     open,
			High:     high,
			Low:      low,
			Close:    close,
			Volume:   vol,
		})
	}

	return model.Response[[]model.Candle]{
		Kind:     model.KindCandle,
		Provider: providerID,
		Market:   market,
		Data:     candles,
	}, nil
}

// OrderBook fetches the order book (depth snapshot) for the given symbol.
func (c *Client) OrderBook(_ context.Context, symbol string, market model.MarketType, depth int) (model.Response[model.OrderBook], error) {
	if market == model.MarketFutures {
		symbol = translateFuturesSymbol(symbol)
	}
	path := fmt.Sprintf("/api/v4/public/orderbook/%s?limit=%d", symbol, depth)
	body, err := c.doRequest(path)
	if err != nil {
		return model.Response[model.OrderBook]{}, err
	}

	var resp whitebitOrderBookResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return model.Response[model.OrderBook]{}, fmt.Errorf("failed to parse orderbook response: %w", err)
	}

	parseEntries := func(raw [][]string) []model.OrderBookEntry {
		entries := make([]model.OrderBookEntry, 0, len(raw))
		for _, e := range raw {
			if len(e) >= 2 {
				price, _ := strconv.ParseFloat(e[0], 64)
				qty, _ := strconv.ParseFloat(e[1], 64)
				entries = append(entries, model.OrderBookEntry{Price: price, Quantity: qty})
			}
		}
		return entries
	}

	bids := parseEntries(resp.Bids)
	asks := parseEntries(resp.Asks)

	var ts *time.Time
	if resp.Timestamp > 0 {
		t := time.Unix(resp.Timestamp, 0)
		ts = &t
	}

	orderbook := model.OrderBook{
		Symbol: symbol,
		Market: market,
		Bids:   bids,
		Asks:   asks,
		Time:   ts,
	}

	return model.Response[model.OrderBook]{
		Kind:     model.KindOrderBook,
		Provider: providerID,
		Market:   market,
		Data:     orderbook,
	}, nil
}

// convertInterval converts a standard interval string to WhiteBit API format.
func (c *Client) futuresTicker24h(ctx context.Context, symbol string, market model.MarketType) (model.Response[model.Ticker24h], error) {
	symbol = translateFuturesSymbol(symbol)
	body, err := c.doRequest("/api/v4/public/futures")
	if err != nil {
		return model.Response[model.Ticker24h]{}, err
	}

	var resp struct {
		Success bool `json:"success"`
		Result  []struct {
			TickerID string `json:"ticker_id"`
			Last     string `json:"last_price"`
			High     string `json:"high_price"`
			Low      string `json:"low_price"`
			Bid      string `json:"bid"`
			Ask      string `json:"ask"`
			Volume   string `json:"base_volume"`
			QuoteVol string `json:"quote_volume"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return model.Response[model.Ticker24h]{}, fmt.Errorf("failed to parse futures ticker: %w", err)
	}

	for _, t := range resp.Result {
		if t.TickerID == symbol {
			lastPrice, _ := strconv.ParseFloat(t.Last, 64)
			highPrice, _ := strconv.ParseFloat(t.High, 64)
			lowPrice, _ := strconv.ParseFloat(t.Low, 64)
			baseVol, _ := strconv.ParseFloat(t.Volume, 64)
			quoteVol, _ := strconv.ParseFloat(t.QuoteVol, 64)
			bidPrice, _ := strconv.ParseFloat(t.Bid, 64)
			askPrice, _ := strconv.ParseFloat(t.Ask, 64)

			return model.Response[model.Ticker24h]{
				Kind:     model.KindTicker,
				Provider: providerID,
				Market:   market,
				Data: model.Ticker24h{
					Symbol:      symbol,
					Market:      market,
					LastPrice:   lastPrice,
					HighPrice:   &highPrice,
					LowPrice:    &lowPrice,
					Volume:      &baseVol,
					QuoteVolume: &quoteVol,
					BidPrice:    &bidPrice,
					AskPrice:    &askPrice,
				},
			}, nil
		}
	}

	return model.Response[model.Ticker24h]{}, fmt.Errorf("ticker not found for symbol %s", symbol)
}

func convertInterval(interval string) string {
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
		return "1h"
	case "4h":
		return "4h"
	case "1d":
		return "1d"
	case "1w":
		return "1w"
	default:
		return interval
	}
}
