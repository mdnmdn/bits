package binance

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	goBinance "github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/mdnmdn/bits/internal/model"
)

// Price fetches the current price for each symbol in ids.
// currency is ignored for exchange providers (symbols encode the quote asset).
func (c *Client) Price(ctx context.Context, ids []string, currency string) (model.Response[[]model.CoinPrice], error) {
	prices := make([]model.CoinPrice, 0, len(ids))
	var errs []model.ItemError

	for _, id := range ids {
		sym := strings.ToUpper(id)
		cp, err := c.fetchPrice(ctx, sym)
		if err != nil {
			errs = append(errs, model.ItemError{Symbol: sym, Err: err})
			continue
		}
		prices = append(prices, *cp)
	}

	return model.Response[[]model.CoinPrice]{
		Provider: providerID,
		Data:     prices,
		Errors:   errs,
	}, nil
}

func (c *Client) fetchPrice(ctx context.Context, sym string) (*model.CoinPrice, error) {
	var priceStr string
	var changePctStr string

	if c.spotClient != nil {
		tickerPrices, err := c.spotClient.NewListPricesService().Symbol(sym).Do(ctx)
		if err != nil {
			return nil, fmt.Errorf("price for %s: %w", sym, err)
		}
		if len(tickerPrices) == 0 {
			return nil, fmt.Errorf("no price data for %s", sym)
		}
		priceStr = tickerPrices[0].Price

		stats, err := c.spotClient.NewListPriceChangeStatsService().Symbol(sym).Do(ctx)
		if err == nil && len(stats) > 0 {
			changePctStr = stats[0].PriceChangePercent
		}
	} else if c.futuresClient != nil {
		tickerPrices, err := c.futuresClient.NewListPricesService().Symbol(sym).Do(ctx)
		if err != nil {
			return nil, fmt.Errorf("futures price for %s: %w", sym, err)
		}
		if len(tickerPrices) == 0 {
			return nil, fmt.Errorf("no futures price data for %s", sym)
		}
		priceStr = tickerPrices[0].Price

		stats, err := c.futuresClient.NewListPriceChangeStatsService().Symbol(sym).Do(ctx)
		if err == nil && len(stats) > 0 {
			changePctStr = stats[0].PriceChangePercent
		}
	} else {
		return nil, fmt.Errorf("binance: no client configured")
	}

	price, _ := strconv.ParseFloat(priceStr, 64)
	cp := &model.CoinPrice{
		ID:     sym,
		Symbol: sym,
		Price:  price,
	}

	if changePctStr != "" {
		v, err := strconv.ParseFloat(changePctStr, 64)
		if err == nil {
			cp.Change24h = &v
		}
	}

	return cp, nil
}

// Candles fetches OHLCV klines for symbol on the given market.
func (c *Client) Candles(ctx context.Context, symbol string, market model.MarketType, interval string, opts model.CandleOpts) (model.Response[[]model.Candle], error) {
	sym := strings.ToUpper(symbol)

	var candles []model.Candle
	var err error

	switch market {
	case model.MarketFutures:
		if c.futuresClient == nil {
			return model.Response[[]model.Candle]{}, fmt.Errorf("binance: futures client not configured")
		}
		candles, err = c.fetchFuturesCandles(ctx, sym, interval, opts)
	default:
		if c.spotClient == nil {
			return model.Response[[]model.Candle]{}, fmt.Errorf("binance: spot client not configured")
		}
		candles, err = c.fetchSpotCandles(ctx, sym, interval, opts)
	}

	if err != nil {
		return model.Response[[]model.Candle]{}, err
	}

	return model.Response[[]model.Candle]{
		Provider: providerID,
		Market:   market,
		Data:     candles,
	}, nil
}

func (c *Client) fetchSpotCandles(ctx context.Context, sym, interval string, opts model.CandleOpts) ([]model.Candle, error) {
	svc := c.spotClient.NewKlinesService().Symbol(sym).Interval(interval)

	if opts.From != nil {
		svc = svc.StartTime(opts.From.UnixMilli())
	}
	if opts.To != nil {
		svc = svc.EndTime(opts.To.UnixMilli())
	}
	if opts.Limit != nil {
		svc = svc.Limit(*opts.Limit)
	}

	klines, err := svc.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("binance: spot candles for %s: %w", sym, err)
	}

	return convertKlines(klines), nil
}

func (c *Client) fetchFuturesCandles(ctx context.Context, sym, interval string, opts model.CandleOpts) ([]model.Candle, error) {
	svc := c.futuresClient.NewKlinesService().Symbol(sym).Interval(interval)

	if opts.From != nil {
		svc = svc.StartTime(opts.From.UnixMilli())
	}
	if opts.To != nil {
		svc = svc.EndTime(opts.To.UnixMilli())
	}
	if opts.Limit != nil {
		svc = svc.Limit(*opts.Limit)
	}

	klines, err := svc.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("binance: futures candles for %s: %w", sym, err)
	}

	return convertFuturesKlines(klines), nil
}

func convertKlines(klines []*goBinance.Kline) []model.Candle {
	result := make([]model.Candle, 0, len(klines))
	for _, k := range klines {
		open, _ := strconv.ParseFloat(k.Open, 64)
		high, _ := strconv.ParseFloat(k.High, 64)
		low, _ := strconv.ParseFloat(k.Low, 64)
		close_, _ := strconv.ParseFloat(k.Close, 64)
		vol, _ := strconv.ParseFloat(k.Volume, 64)
		ct := time.UnixMilli(k.CloseTime)
		result = append(result, model.Candle{
			OpenTime:  time.UnixMilli(k.OpenTime),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close_,
			Volume:    &vol,
			CloseTime: &ct,
		})
	}
	return result
}

func convertFuturesKlines(klines []*futures.Kline) []model.Candle {
	result := make([]model.Candle, 0, len(klines))
	for _, k := range klines {
		open, _ := strconv.ParseFloat(k.Open, 64)
		high, _ := strconv.ParseFloat(k.High, 64)
		low, _ := strconv.ParseFloat(k.Low, 64)
		close_, _ := strconv.ParseFloat(k.Close, 64)
		vol, _ := strconv.ParseFloat(k.Volume, 64)
		ct := time.UnixMilli(k.CloseTime)
		result = append(result, model.Candle{
			OpenTime:  time.UnixMilli(k.OpenTime),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close_,
			Volume:    &vol,
			CloseTime: &ct,
		})
	}
	return result
}

// Ticker24h fetches 24h rolling ticker statistics for symbol on market.
func (c *Client) Ticker24h(ctx context.Context, symbol string, market model.MarketType) (model.Response[model.Ticker24h], error) {
	sym := strings.ToUpper(symbol)

	switch market {
	case model.MarketFutures:
		if c.futuresClient == nil {
			return model.Response[model.Ticker24h]{}, fmt.Errorf("binance: futures client not configured")
		}
		stats, err := c.futuresClient.NewListPriceChangeStatsService().Symbol(sym).Do(ctx)
		if err != nil {
			return model.Response[model.Ticker24h]{}, fmt.Errorf("binance: futures ticker for %s: %w", sym, err)
		}
		if len(stats) == 0 {
			return model.Response[model.Ticker24h]{}, fmt.Errorf("binance: no futures ticker stats for %s", sym)
		}
		s := stats[0]
		t := convertFuturesTicker(s, market)
		return model.Response[model.Ticker24h]{Provider: providerID, Market: market, Data: t}, nil

	default:
		if c.spotClient == nil {
			return model.Response[model.Ticker24h]{}, fmt.Errorf("binance: spot client not configured")
		}
		stats, err := c.spotClient.NewListPriceChangeStatsService().Symbol(sym).Do(ctx)
		if err != nil {
			return model.Response[model.Ticker24h]{}, fmt.Errorf("binance: ticker for %s: %w", sym, err)
		}
		if len(stats) == 0 {
			return model.Response[model.Ticker24h]{}, fmt.Errorf("binance: no ticker stats for %s", sym)
		}
		s := stats[0]
		t := convertSpotTicker(s, market)
		return model.Response[model.Ticker24h]{Provider: providerID, Market: market, Data: t}, nil
	}
}

func convertSpotTicker(s *goBinance.PriceChangeStats, market model.MarketType) model.Ticker24h {
	lastPrice, _ := strconv.ParseFloat(s.LastPrice, 64)
	priceChange, _ := strconv.ParseFloat(s.PriceChange, 64)
	priceChangePct, _ := strconv.ParseFloat(s.PriceChangePercent, 64)
	highPrice, _ := strconv.ParseFloat(s.HighPrice, 64)
	lowPrice, _ := strconv.ParseFloat(s.LowPrice, 64)
	vol, _ := strconv.ParseFloat(s.Volume, 64)
	quoteVol, _ := strconv.ParseFloat(s.QuoteVolume, 64)
	openPrice, _ := strconv.ParseFloat(s.OpenPrice, 64)
	weightedAvg, _ := strconv.ParseFloat(s.WeightedAvgPrice, 64)
	bidPrice, _ := strconv.ParseFloat(s.BidPrice, 64)
	askPrice, _ := strconv.ParseFloat(s.AskPrice, 64)
	ot := time.UnixMilli(s.OpenTime)
	ct := time.UnixMilli(s.CloseTime)

	return model.Ticker24h{
		Symbol:             s.Symbol,
		Market:             market,
		LastPrice:          lastPrice,
		PriceChange:        &priceChange,
		PriceChangePercent: &priceChangePct,
		HighPrice:          &highPrice,
		LowPrice:           &lowPrice,
		Volume:             &vol,
		QuoteVolume:        &quoteVol,
		OpenPrice:          &openPrice,
		WeightedAvgPrice:   &weightedAvg,
		BidPrice:           &bidPrice,
		AskPrice:           &askPrice,
		OpenTime:           &ot,
		CloseTime:          &ct,
	}
}

func convertFuturesTicker(s *futures.PriceChangeStats, market model.MarketType) model.Ticker24h {
	lastPrice, _ := strconv.ParseFloat(s.LastPrice, 64)
	priceChange, _ := strconv.ParseFloat(s.PriceChange, 64)
	priceChangePct, _ := strconv.ParseFloat(s.PriceChangePercent, 64)
	highPrice, _ := strconv.ParseFloat(s.HighPrice, 64)
	lowPrice, _ := strconv.ParseFloat(s.LowPrice, 64)
	vol, _ := strconv.ParseFloat(s.Volume, 64)
	quoteVol, _ := strconv.ParseFloat(s.QuoteVolume, 64)
	openPrice, _ := strconv.ParseFloat(s.OpenPrice, 64)
	weightedAvg, _ := strconv.ParseFloat(s.WeightedAvgPrice, 64)
	ot := time.UnixMilli(s.OpenTime)
	ct := time.UnixMilli(s.CloseTime)

	return model.Ticker24h{
		Symbol:             s.Symbol,
		Market:             market,
		LastPrice:          lastPrice,
		PriceChange:        &priceChange,
		PriceChangePercent: &priceChangePct,
		HighPrice:          &highPrice,
		LowPrice:           &lowPrice,
		Volume:             &vol,
		QuoteVolume:        &quoteVol,
		OpenPrice:          &openPrice,
		WeightedAvgPrice:   &weightedAvg,
		OpenTime:           &ot,
		CloseTime:          &ct,
	}
}

// OrderBook fetches the order book depth snapshot for symbol on market.
func (c *Client) OrderBook(ctx context.Context, symbol string, market model.MarketType, depth int) (model.Response[model.OrderBook], error) {
	sym := strings.ToUpper(symbol)
	if depth <= 0 {
		depth = 20
	}

	switch market {
	case model.MarketFutures:
		if c.futuresClient == nil {
			return model.Response[model.OrderBook]{}, fmt.Errorf("binance: futures client not configured")
		}
		d, err := c.futuresClient.NewDepthService().Symbol(sym).Limit(depth).Do(ctx)
		if err != nil {
			return model.Response[model.OrderBook]{}, fmt.Errorf("binance: futures order book for %s: %w", sym, err)
		}
		uid := d.LastUpdateID
		ob := model.OrderBook{
			Symbol:       sym,
			Market:       market,
			LastUpdateID: &uid,
			Bids:         convertFuturesDepth(d.Bids),
			Asks:         convertFuturesDepth(d.Asks),
		}
		return model.Response[model.OrderBook]{Provider: providerID, Market: market, Data: ob}, nil

	default:
		if c.spotClient == nil {
			return model.Response[model.OrderBook]{}, fmt.Errorf("binance: spot client not configured")
		}
		d, err := c.spotClient.NewDepthService().Symbol(sym).Limit(depth).Do(ctx)
		if err != nil {
			return model.Response[model.OrderBook]{}, fmt.Errorf("binance: order book for %s: %w", sym, err)
		}
		uid := d.LastUpdateID
		ob := model.OrderBook{
			Symbol:       sym,
			Market:       market,
			LastUpdateID: &uid,
			Bids:         convertSpotDepth(d.Bids),
			Asks:         convertSpotDepth(d.Asks),
		}
		return model.Response[model.OrderBook]{Provider: providerID, Market: market, Data: ob}, nil
	}
}

func convertSpotDepth(entries []goBinance.Bid) []model.OrderBookEntry {
	result := make([]model.OrderBookEntry, 0, len(entries))
	for _, e := range entries {
		price, _ := strconv.ParseFloat(e.Price, 64)
		qty, _ := strconv.ParseFloat(e.Quantity, 64)
		result = append(result, model.OrderBookEntry{Price: price, Quantity: qty})
	}
	return result
}

func convertFuturesDepth(entries []futures.Bid) []model.OrderBookEntry {
	result := make([]model.OrderBookEntry, 0, len(entries))
	for _, e := range entries {
		price, _ := strconv.ParseFloat(e.Price, 64)
		qty, _ := strconv.ParseFloat(e.Quantity, 64)
		result = append(result, model.OrderBookEntry{Price: price, Quantity: qty})
	}
	return result
}
