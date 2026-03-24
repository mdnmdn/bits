package bitget

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/coingecko/coingecko-cli/internal/model"
)

// SimplePrice implements the Provider interface.
// ids are Bitget symbols (e.g. "BTCUSDT"). For each symbol it calls the ticker
// endpoint and returns the close price and 24h change percentage.
func (c *Client) SimplePrice(_ context.Context, ids []string, vsCurrency string) (model.PriceResponse, error) {
	result := make(model.PriceResponse, len(ids))

	for _, symbol := range ids {
		ticker, err := c.GetTickerData(symbol)
		if err != nil {
			return nil, fmt.Errorf("failed to get ticker for %s: %w", symbol, err)
		}

		price, err := strconv.ParseFloat(ticker.Close, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse price for %s: %w", symbol, err)
		}

		changePct, _ := strconv.ParseFloat(ticker.ChangeUtc, 64)
		// ChangeUtc is a ratio (e.g. 0.05 = 5%), convert to percentage
		changePct *= 100

		result[symbol] = map[string]float64{
			vsCurrency:                          price,
			vsCurrency + "_24h_change":          changePct,
		}
	}

	return result, nil
}

// CoinOHLC implements the Provider interface.
// It fetches historical candle data and returns model.OHLCData.
func (c *Client) CoinOHLC(_ context.Context, id, _ string, days, interval string) (model.OHLCData, error) {
	// Map interval to Bitget granularity
	granularity := interval
	if granularity == "" {
		granularity = "daily"
	}

	// Calculate time range from days
	var startTime int64
	endTime := time.Now().UnixMilli()

	if days != "" {
		d, err := strconv.Atoi(days)
		if err != nil {
			return nil, fmt.Errorf("invalid days value %q: %w", days, err)
		}
		startTime = time.Now().Add(-time.Duration(d) * 24 * time.Hour).UnixMilli()
	}

	candles, err := c.GetHistoricalCandles(id, granularity, startTime, endTime, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get candles for %s: %w", id, err)
	}

	// Convert to model.OHLCData: each entry is [timestamp_ms, open, high, low, close]
	ohlc := make(model.OHLCData, 0, len(candles))
	for _, candle := range candles {
		ohlc = append(ohlc, []float64{
			float64(candle.OpenTime),
			candle.OpenPrice,
			candle.HighPrice,
			candle.LowPrice,
			candle.ClosePrice,
		})
	}

	return ohlc, nil
}

// Ticker24h implements the TickerProvider interface.
// It returns 24-hour ticker statistics for the given symbol.
func (c *Client) Ticker24h(_ context.Context, symbol string) (*model.Ticker24h, error) {
	ticker, err := c.GetTickerData(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticker for %s: %w", symbol, err)
	}

	lastPrice, _ := strconv.ParseFloat(ticker.Close, 64)
	highPrice, _ := strconv.ParseFloat(ticker.High24h, 64)
	lowPrice, _ := strconv.ParseFloat(ticker.Low24h, 64)
	openPrice, _ := strconv.ParseFloat(ticker.OpenUtc0, 64)
	volume, _ := strconv.ParseFloat(ticker.BaseVol, 64)
	quoteVolume, _ := strconv.ParseFloat(ticker.QuoteVol, 64)
	changePct, _ := strconv.ParseFloat(ticker.ChangeUtc, 64)
	changePct *= 100 // Convert ratio to percentage

	priceChange := lastPrice - openPrice

	ts, _ := strconv.ParseInt(ticker.Ts, 10, 64)
	closeTime := time.UnixMilli(ts)
	openTime := closeTime.Add(-24 * time.Hour)

	return &model.Ticker24h{
		Symbol:             ticker.Symbol,
		LastPrice:          lastPrice,
		PriceChange:        priceChange,
		PriceChangePercent: changePct,
		HighPrice:          highPrice,
		LowPrice:           lowPrice,
		Volume:             volume,
		QuoteVolume:        quoteVolume,
		OpenPrice:          openPrice,
		OpenTime:           openTime,
		CloseTime:          closeTime,
	}, nil
}

// GetTickerData fetches detailed ticker data for a symbol.
func (c *Client) GetTickerData(symbol string) (*TickerData, error) {
	endpoint := "/api/spot/v1/market/ticker"
	query := "symbol=" + symbol

	body, err := c.signedRequest("GET", endpoint, query, nil)
	if err != nil {
		return nil, err
	}

	var ticker TickerResponse
	if err := json.Unmarshal(body, &ticker); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w, body: %s", err, string(body))
	}

	if ticker.Code != "00000" {
		return nil, fmt.Errorf("API error: %s", ticker.Msg)
	}

	return &ticker.Data, nil
}

// GetHistoricalCandles gets historical candle data for a trading pair.
func (c *Client) GetHistoricalCandles(symbol, granularity string, startTime, endTime int64, limit int) ([]CandleData, error) {
	endpoint := "/api/v2/spot/market/history-candles"

	apiGranularity := convertGranularityFormat(granularity)

	params := []string{
		fmt.Sprintf("symbol=%s", symbol),
		fmt.Sprintf("granularity=%s", apiGranularity),
	}

	if startTime > 0 {
		params = append(params, fmt.Sprintf("startTime=%d", startTime))
	}
	if endTime > 0 {
		params = append(params, fmt.Sprintf("endTime=%d", endTime))
	}
	if limit > 0 {
		params = append(params, fmt.Sprintf("limit=%d", limit))
	}

	queryString := strings.Join(params, "&")

	body, err := c.signedRequest("GET", endpoint, queryString, nil)
	if err != nil {
		return nil, err
	}

	var candleResp HistoricalCandlesResponse
	if err := json.Unmarshal(body, &candleResp); err != nil {
		return nil, fmt.Errorf("failed to parse candle response: %w, body: %s", err, string(body))
	}

	if candleResp.Code != "00000" {
		return nil, fmt.Errorf("API error: %s", candleResp.Msg)
	}

	candles := make([]CandleData, len(candleResp.Data))
	for i, candleArray := range candleResp.Data {
		if len(candleArray) < 6 {
			return nil, fmt.Errorf("invalid candle data format at index %d: expected at least 6 fields, got %d", i, len(candleArray))
		}

		candle := CandleData{
			Timestamp: candleArray[0],
			Open:      candleArray[1],
			High:      candleArray[2],
			Low:       candleArray[3],
			Close:     candleArray[4],
			Volume:    candleArray[5],
		}

		if timestampInt, err := strconv.ParseInt(candle.Timestamp, 10, 64); err == nil {
			candle.OpenTime = timestampInt
			candle.CloseTime = timestampInt
		}
		if v, err := strconv.ParseFloat(candle.Open, 64); err == nil {
			candle.OpenPrice = v
		}
		if v, err := strconv.ParseFloat(candle.High, 64); err == nil {
			candle.HighPrice = v
		}
		if v, err := strconv.ParseFloat(candle.Low, 64); err == nil {
			candle.LowPrice = v
		}
		if v, err := strconv.ParseFloat(candle.Close, 64); err == nil {
			candle.ClosePrice = v
		}
		if v, err := strconv.ParseFloat(candle.Volume, 64); err == nil {
			candle.VolumeFloat = v
		}

		candles[i] = candle
	}

	return candles, nil
}
