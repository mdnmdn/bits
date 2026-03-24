package binance

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mdnmdn/bits/internal/model"
)

// SimplePrice returns current prices for the given Binance symbols.
// ids are Binance trading pair symbols (e.g. "BTCUSDT").
// vsCurrency is used as the key in the inner price map.
func (c *Client) SimplePrice(ctx context.Context, ids []string, vsCurrency string) (model.PriceResponse, error) {
	result := make(model.PriceResponse, len(ids))
	vs := strings.ToLower(vsCurrency)

	for _, symbol := range ids {
		sym := strings.ToUpper(symbol)

		prices, err := c.client.NewListPricesService().Symbol(sym).Do(ctx)
		if err != nil {
			return nil, fmt.Errorf("fetching price for %s: %w", sym, err)
		}
		if len(prices) == 0 {
			return nil, fmt.Errorf("no price data found for symbol %s", sym)
		}

		price, err := strconv.ParseFloat(prices[0].Price, 64)
		if err != nil {
			return nil, fmt.Errorf("parsing price for %s: %w", sym, err)
		}

		stats, err := c.client.NewListPriceChangeStatsService().Symbol(sym).Do(ctx)
		if err != nil {
			return nil, fmt.Errorf("fetching 24h stats for %s: %w", sym, err)
		}

		inner := map[string]float64{
			vs: price,
		}

		if len(stats) > 0 {
			changePct, err := strconv.ParseFloat(stats[0].PriceChangePercent, 64)
			if err == nil {
				inner[vs+"_24h_change"] = changePct
			}
		}

		result[sym] = inner
	}

	return result, nil
}

// CoinOHLC returns OHLC candlestick data for a Binance symbol.
// days is the number of days of history (as a string). interval maps:
// "daily" -> "1d", "hourly" -> "1h", otherwise pass-through; default is "1d".
func (c *Client) CoinOHLC(ctx context.Context, id, vsCurrency, days, interval string) (model.OHLCData, error) {
	sym := strings.ToUpper(id)

	// Map interval to Binance kline interval
	binanceInterval := "1d"
	switch strings.ToLower(interval) {
	case "daily", "":
		binanceInterval = "1d"
	case "hourly":
		binanceInterval = "1h"
	default:
		binanceInterval = interval
	}

	// Parse days to compute start time
	numDays, err := strconv.Atoi(days)
	if err != nil {
		numDays = 30 // default fallback
	}

	now := time.Now()
	startTime := now.AddDate(0, 0, -numDays).UnixMilli()
	endTime := now.UnixMilli()

	klines, err := c.client.NewKlinesService().
		Symbol(sym).
		Interval(binanceInterval).
		StartTime(startTime).
		EndTime(endTime).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching klines for %s: %w", sym, err)
	}

	data := make(model.OHLCData, 0, len(klines))
	for _, k := range klines {
		open, _ := strconv.ParseFloat(k.Open, 64)
		high, _ := strconv.ParseFloat(k.High, 64)
		low, _ := strconv.ParseFloat(k.Low, 64)
		closePrice, _ := strconv.ParseFloat(k.Close, 64)

		data = append(data, []float64{
			float64(k.OpenTime),
			open,
			high,
			low,
			closePrice,
		})
	}

	return data, nil
}

// Ticker24h returns 24-hour ticker statistics for a Binance symbol.
func (c *Client) Ticker24h(ctx context.Context, symbol string) (*model.Ticker24h, error) {
	sym := strings.ToUpper(symbol)

	stats, err := c.client.NewListPriceChangeStatsService().Symbol(sym).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching 24h ticker for %s: %w", sym, err)
	}
	if len(stats) == 0 {
		return nil, fmt.Errorf("no ticker stats found for symbol %s", sym)
	}

	s := stats[0]

	lastPrice, _ := strconv.ParseFloat(s.LastPrice, 64)
	priceChange, _ := strconv.ParseFloat(s.PriceChange, 64)
	priceChangePct, _ := strconv.ParseFloat(s.PriceChangePercent, 64)
	highPrice, _ := strconv.ParseFloat(s.HighPrice, 64)
	lowPrice, _ := strconv.ParseFloat(s.LowPrice, 64)
	volume, _ := strconv.ParseFloat(s.Volume, 64)
	quoteVolume, _ := strconv.ParseFloat(s.QuoteVolume, 64)
	openPrice, _ := strconv.ParseFloat(s.OpenPrice, 64)
	weightedAvg, _ := strconv.ParseFloat(s.WeightedAvgPrice, 64)

	return &model.Ticker24h{
		Symbol:             s.Symbol,
		LastPrice:          lastPrice,
		PriceChange:        priceChange,
		PriceChangePercent: priceChangePct,
		HighPrice:          highPrice,
		LowPrice:           lowPrice,
		Volume:             volume,
		QuoteVolume:        quoteVolume,
		OpenPrice:          openPrice,
		WeightedAvgPrice:   weightedAvg,
		OpenTime:           time.UnixMilli(s.OpenTime),
		CloseTime:          time.UnixMilli(s.CloseTime),
	}, nil
}

// OrderBook returns the order book depth for a Binance symbol.
func (c *Client) OrderBook(ctx context.Context, symbol string, limit int) (*model.OrderBook, error) {
	sym := strings.ToUpper(symbol)

	if limit <= 0 {
		limit = 100
	}

	depth, err := c.client.NewDepthService().
		Symbol(sym).
		Limit(limit).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching order book for %s: %w", sym, err)
	}

	bids := make([]model.OrderBookEntry, 0, len(depth.Bids))
	for _, bid := range depth.Bids {
		price, _ := strconv.ParseFloat(bid.Price, 64)
		qty, _ := strconv.ParseFloat(bid.Quantity, 64)
		bids = append(bids, model.OrderBookEntry{
			Price:    price,
			Quantity: qty,
		})
	}

	asks := make([]model.OrderBookEntry, 0, len(depth.Asks))
	for _, ask := range depth.Asks {
		price, _ := strconv.ParseFloat(ask.Price, 64)
		qty, _ := strconv.ParseFloat(ask.Quantity, 64)
		asks = append(asks, model.OrderBookEntry{
			Price:    price,
			Quantity: qty,
		})
	}

	return &model.OrderBook{
		Symbol: sym,
		Bids:   bids,
		Asks:   asks,
	}, nil
}
