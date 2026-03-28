package bitget

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/mdnmdn/bits/internal/model"
)

// bitgetTickerEntry represents a single ticker item in the Bitget v2 response.
// Used for both spot and futures ticker endpoints.
type bitgetTickerEntry struct {
	Symbol    string `json:"symbol"`
	LastPr    string `json:"lastPr"`
	High24H   string `json:"high24h"`
	Low24H    string `json:"low24h"`
	Open      string `json:"open"`    // spot open price
	OpenUtc   string `json:"openUtc"` // futures open price at UTC 0
	BaseVol   string `json:"baseVolume"`
	QuoteVol  string `json:"quoteVolume"`
	BidPr     string `json:"bidPr"`
	AskPr     string `json:"askPr"`
	Change24H string `json:"change24h"` // ratio, e.g. 0.02 = 2%
	Ts        string `json:"ts"`
}

// bitgetTickersResponse is the response envelope for spot/futures ticker endpoints.
type bitgetTickersResponse struct {
	Code string              `json:"code"`
	Msg  string              `json:"msg"`
	Data []bitgetTickerEntry `json:"data"`
}

// bitgetCandlesResponse is the response envelope for candle endpoints.
type bitgetCandlesResponse struct {
	Code string     `json:"code"`
	Msg  string     `json:"msg"`
	Data [][]string `json:"data"`
}

// Price fetches current prices for the given symbols.
// ids are trading symbols (e.g. "BTCUSDT"). currency is used as metadata only.
func (c *Client) Price(_ context.Context, ids []string, currency string) (model.Response[[]model.CoinPrice], error) {
	prices := make([]model.CoinPrice, 0, len(ids))
	var itemErrors []model.ItemError

	for _, symbol := range ids {
		entry, err := c.fetchTicker(symbol, model.MarketSpot)
		if err != nil {
			itemErrors = append(itemErrors, model.ItemError{Symbol: symbol, Err: err})
			continue
		}

		price, _ := strconv.ParseFloat(entry.LastPr, 64)
		changePct, _ := strconv.ParseFloat(entry.Change24H, 64)
		changePct *= 100 // Convert ratio to percentage

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
		Data:     prices,
		Errors:   itemErrors,
	}, nil
}

// Candles fetches OHLCV candle data for the given symbol and market.
func (c *Client) Candles(_ context.Context, symbol string, market model.MarketType, interval string, opts model.CandleOpts) (model.Response[[]model.Candle], error) {
	var path string
	var granularity string

	switch market {
	case model.MarketFutures:
		path = "/api/v2/mix/market/history-candles"
		granularity = convertGranularityFutures(interval)
	default:
		granularity = convertGranularitySpot(interval)
		// history-candles requires startTime; use the plain candles endpoint when
		// no time range is requested so that limit-based fetching works.
		if opts.From != nil {
			path = "/api/v2/spot/market/history-candles"
		} else {
			path = "/api/v2/spot/market/candles"
		}
	}

	query := fmt.Sprintf("symbol=%s&granularity=%s", symbol, granularity)

	if market == model.MarketFutures {
		query += "&productType=USDT-FUTURES"
	}

	if opts.From != nil {
		query += fmt.Sprintf("&startTime=%d", opts.From.UnixMilli())
	}
	if opts.To != nil {
		query += fmt.Sprintf("&endTime=%d", opts.To.UnixMilli())
	}
	if opts.Limit != nil {
		query += fmt.Sprintf("&limit=%d", *opts.Limit)
	}

	body, err := c.doRequest("GET", path, query)
	if err != nil {
		return model.Response[[]model.Candle]{}, err
	}

	var resp bitgetCandlesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return model.Response[[]model.Candle]{}, fmt.Errorf("failed to parse candles response: %w", err)
	}
	if resp.Code != "00000" {
		return model.Response[[]model.Candle]{}, fmt.Errorf("API error: %s", resp.Msg)
	}

	candles := make([]model.Candle, 0, len(resp.Data))
	for i, row := range resp.Data {
		if len(row) < 6 {
			return model.Response[[]model.Candle]{}, fmt.Errorf("invalid candle data at index %d: expected at least 6 fields, got %d", i, len(row))
		}

		tsMs, _ := strconv.ParseInt(row[0], 10, 64)
		openTime := time.UnixMilli(tsMs)
		open, _ := strconv.ParseFloat(row[1], 64)
		high, _ := strconv.ParseFloat(row[2], 64)
		low, _ := strconv.ParseFloat(row[3], 64)
		close, _ := strconv.ParseFloat(row[4], 64)
		vol, _ := strconv.ParseFloat(row[5], 64)

		candles = append(candles, model.Candle{
			OpenTime: openTime,
			Open:     open,
			High:     high,
			Low:      low,
			Close:    close,
			Volume:   &vol,
		})
	}

	return model.Response[[]model.Candle]{
		Kind:     model.KindCandle,
		Provider: providerID,
		Market:   market,
		Data:     candles,
	}, nil
}

// Ticker24h fetches 24-hour rolling ticker statistics for the given symbol and market.
func (c *Client) Ticker24h(_ context.Context, symbol string, market model.MarketType) (model.Response[model.Ticker24h], error) {
	entry, err := c.fetchTicker(symbol, market)
	if err != nil {
		return model.Response[model.Ticker24h]{}, err
	}

	lastPrice, _ := strconv.ParseFloat(entry.LastPr, 64)
	highPrice, _ := strconv.ParseFloat(entry.High24H, 64)
	lowPrice, _ := strconv.ParseFloat(entry.Low24H, 64)
	baseVol, _ := strconv.ParseFloat(entry.BaseVol, 64)
	quoteVol, _ := strconv.ParseFloat(entry.QuoteVol, 64)
	bidPrice, _ := strconv.ParseFloat(entry.BidPr, 64)
	askPrice, _ := strconv.ParseFloat(entry.AskPr, 64)
	changePct, _ := strconv.ParseFloat(entry.Change24H, 64)
	changePct *= 100 // Convert ratio to percentage

	// Determine open price based on market
	openStr := entry.Open
	if market == model.MarketFutures {
		openStr = entry.OpenUtc
	}
	openPrice, _ := strconv.ParseFloat(openStr, 64)
	priceChange := lastPrice - openPrice

	var closeTime *time.Time
	var openTime *time.Time
	if ts, err := strconv.ParseInt(entry.Ts, 10, 64); err == nil {
		ct := time.UnixMilli(ts)
		ot := ct.Add(-24 * time.Hour)
		closeTime = &ct
		openTime = &ot
	}

	ticker := model.Ticker24h{
		Symbol:             symbol,
		Market:             market,
		LastPrice:          lastPrice,
		PriceChange:        &priceChange,
		PriceChangePercent: &changePct,
		HighPrice:          &highPrice,
		LowPrice:           &lowPrice,
		Volume:             &baseVol,
		QuoteVolume:        &quoteVol,
		OpenPrice:          &openPrice,
		BidPrice:           &bidPrice,
		AskPrice:           &askPrice,
		OpenTime:           openTime,
		CloseTime:          closeTime,
	}

	return model.Response[model.Ticker24h]{
		Kind:     model.KindTicker,
		Provider: providerID,
		Market:   market,
		Data:     ticker,
	}, nil
}

// fetchTicker calls the appropriate ticker endpoint and returns the first entry.
func (c *Client) fetchTicker(symbol string, market model.MarketType) (*bitgetTickerEntry, error) {
	var path, query string

	switch market {
	case model.MarketFutures:
		path = "/api/v2/mix/market/ticker"
		query = fmt.Sprintf("symbol=%s&productType=USDT-FUTURES", symbol)
	default:
		path = "/api/v2/spot/market/tickers"
		query = fmt.Sprintf("symbol=%s", symbol)
	}

	body, err := c.doRequest("GET", path, query)
	if err != nil {
		return nil, err
	}

	var resp bitgetTickersResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse ticker response: %w", err)
	}
	if resp.Code != "00000" {
		return nil, fmt.Errorf("API error: %s", resp.Msg)
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no ticker data returned for symbol %s", symbol)
	}

	return &resp.Data[0], nil
}

// convertGranularitySpot converts an interval string to Bitget spot API granularity format.
func convertGranularitySpot(interval string) string {
	switch interval {
	case "1m":
		return "1min"
	case "5m":
		return "5min"
	case "15m":
		return "15min"
	case "30m":
		return "30min"
	case "1h":
		return "1h"
	case "4h":
		return "4h"
	case "1d":
		return "1day"
	default:
		return interval
	}
}

// convertGranularityFutures converts an interval string to Bitget futures API granularity format.
func convertGranularityFutures(interval string) string {
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
		return "1H"
	case "4h":
		return "4H"
	case "1d":
		return "1D"
	default:
		return interval
	}
}
