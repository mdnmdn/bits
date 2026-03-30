package cryptocom

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/mdnmdn/bits/pkg/model"
)

// Price fetches current prices for the given symbols.
// ids are trading pair symbols in Crypto.com format (e.g. "BTC_USDT").
// currency is used as metadata only; Crypto.com ticker does not filter by quote currency.
func (c *Client) Price(_ context.Context, ids []string, currency string) (model.Response[[]model.CoinPrice], error) {
	prices := make([]model.CoinPrice, 0, len(ids))
	var itemErrors []model.ItemError

	for _, symbol := range ids {
		data, err := c.fetchTicker(symbol)
		if err != nil {
			itemErrors = append(itemErrors, model.ItemError{Symbol: symbol, Err: err})
			continue
		}

		price, _ := strconv.ParseFloat(data.A, 64)
		changePctStr, _ := strconv.ParseFloat(data.C, 64)

		var changePct *float64
		if changePctStr != 0 {
			v := changePctStr * 100
			changePct = &v
		}

		prices = append(prices, model.CoinPrice{
			ID:        symbol,
			Symbol:    symbol,
			Currency:  currency,
			Price:     price,
			Change24h: changePct,
		})
	}

	return model.Response[[]model.CoinPrice]{
		Kind:     model.KindPrice,
		Provider: providerID,
		Market:   model.MarketSpot,
		Data:     prices,
		Errors:   itemErrors,
	}, nil
}

// Ticker24h fetches 24-hour rolling ticker statistics for the given symbol.
// Only spot market is supported; the market parameter is accepted for interface compatibility.
func (c *Client) Ticker24h(_ context.Context, symbol string, market model.MarketType) (model.Response[model.Ticker24h], error) {
	data, err := c.fetchTicker(symbol)
	if err != nil {
		return model.Response[model.Ticker24h]{}, err
	}

	lastPrice, _ := strconv.ParseFloat(data.A, 64)
	bidPrice, _ := strconv.ParseFloat(data.B, 64)
	openPrice, _ := strconv.ParseFloat(data.O, 64)
	highPrice, _ := strconv.ParseFloat(data.H, 64)
	lowPrice, _ := strconv.ParseFloat(data.L, 64)
	volume, _ := strconv.ParseFloat(data.V, 64)
	quoteVolume, _ := strconv.ParseFloat(data.VV, 64)
	priceChange, _ := strconv.ParseFloat(data.P, 64)
	priceChangePctStr, _ := strconv.ParseFloat(data.C, 64)

	var priceChangePct *float64
	if priceChangePctStr != 0 {
		v := priceChangePctStr * 100
		priceChangePct = &v
	}

	ticker := model.Ticker24h{
		Symbol:             symbol,
		Market:             market,
		LastPrice:          lastPrice,
		BidPrice:           &bidPrice,
		AskPrice:           &lastPrice,
		OpenPrice:          &openPrice,
		HighPrice:          &highPrice,
		LowPrice:           &lowPrice,
		Volume:             &volume,
		QuoteVolume:        &quoteVolume,
		PriceChange:        &priceChange,
		PriceChangePercent: priceChangePct,
	}

	return model.Response[model.Ticker24h]{
		Kind:     model.KindTicker,
		Provider: providerID,
		Market:   market,
		Data:     ticker,
	}, nil
}

// OrderBook fetches the order book depth snapshot for the given symbol.
// Only spot market is supported.
func (c *Client) OrderBook(_ context.Context, symbol string, market model.MarketType, depth int) (model.Response[model.OrderBook], error) {
	query := fmt.Sprintf("instrument_name=%s", symbol)
	if depth > 0 {
		query += fmt.Sprintf("&depth=%d", depth)
	}

	body, err := c.doRequest("public/get-book", query)
	if err != nil {
		return model.Response[model.OrderBook]{}, err
	}

	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return model.Response[model.OrderBook]{}, fmt.Errorf("failed to parse book response: %w", err)
	}
	code, err := env.GetCode()
	if err != nil {
		return model.Response[model.OrderBook]{}, fmt.Errorf("failed to parse code: %w", err)
	}
	if code != 0 {
		return model.Response[model.OrderBook]{}, fmt.Errorf("API error (code %d): %s", code, env.Msg)
	}

	var result apiBookResult
	if err := json.Unmarshal(env.Result, &result); err != nil {
		return model.Response[model.OrderBook]{}, fmt.Errorf("failed to parse book result: %w", err)
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

	var bids, asks [][]string
	var snapshotTime *time.Time

	if len(result.Data) > 0 {
		row := result.Data[0]
		bids = row.Bids
		asks = row.Asks
		if row.T > 0 {
			t := time.UnixMilli(row.T)
			snapshotTime = &t
		}
	}
	if snapshotTime == nil {
		now := time.Now()
		snapshotTime = &now
	}

	orderbook := model.OrderBook{
		Symbol: symbol,
		Market: market,
		Bids:   parseEntries(bids),
		Asks:   parseEntries(asks),
		Time:   snapshotTime,
	}

	return model.Response[model.OrderBook]{
		Kind:     model.KindOrderBook,
		Provider: providerID,
		Market:   market,
		Data:     orderbook,
	}, nil
}

// fetchTicker calls public/get-ticker for a single instrument and returns the first data entry.
func (c *Client) fetchTicker(symbol string) (*apiTickerData, error) {
	query := fmt.Sprintf("instrument_name=%s", symbol)
	body, err := c.doRequest("public/get-ticker", query)
	if err != nil {
		return nil, err
	}

	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("failed to parse ticker response: %w", err)
	}
	code, err := env.GetCode()
	if err != nil {
		return nil, fmt.Errorf("failed to parse code: %w", err)
	}
	if code != 0 {
		return nil, fmt.Errorf("API error (code %d): %s", code, env.Msg)
	}

	var result apiTickerResult
	if err := json.Unmarshal(env.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse ticker result: %w", err)
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no ticker data for symbol %s", symbol)
	}

	return &result.Data[0], nil
}
