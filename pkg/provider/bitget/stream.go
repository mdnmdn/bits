package bitget

import (
	"context"
	"time"

	"github.com/mdnmdn/bits/internal/logger"
	"github.com/mdnmdn/bits/internal/ws"
	"github.com/mdnmdn/bits/internal/ws/middleware"
	"github.com/mdnmdn/bits/pkg/model"
)

const wsURL = "wss://ws.bitget.com/v2/ws/public"

// pool creates a new WebSocket pool for this client.
func (c *Client) pool(ctx context.Context) (*ws.Pool, <-chan ws.StreamResponse[any], error) {
	cfg := ws.SessionConfig{
		ConnConfig: ws.ConnConfig{
			URL:           wsURL,
			PingInterval:  30 * time.Second,
			OutChanBuffer: 100,
		},
		Protocol: &bitgetProtocol{providerID: providerID},
		Pipeline: ws.Pipeline{
			middleware.CRC32ValidatorMW(),
			middleware.OrderBookReconstructorMW(),
		},
	}
	pool := ws.NewPool(cfg, 0) // 0 = unlimited per connection
	out, err := pool.Start(ctx)
	return pool, out, err
}

// WatchPrices implements provider.PriceStreamProvider.
func (c *Client) WatchPrices(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error) {
	pool, out, err := c.pool(ctx)
	if err != nil {
		return nil, err
	}

	// Subscribe to ticker for each ID
	for _, id := range ids {
		if err := pool.Subscribe(ctx, ws.Subscription{
			Key: "ticker:" + id + ":SPOT",
			Params: []map[string]string{
				{
					"instType": "SPOT",
					"channel":  "ticker",
					"instId":   id,
				},
			},
		}); err != nil {
			pool.Stop()
			return nil, err
		}
	}

	// Extract typed prices
	data, errs := ws.TypedChan[model.CoinPrice](out, 100)

	// Drain error channel in background
	go func() {
		for err := range errs {
			logger.Default.Warn("bitget price stream error", "err", err)
		}
	}()

	// Unwrap Response[CoinPrice] to raw CoinPrice
	prices := make(chan *model.CoinPrice, 100)
	go func() {
		defer close(prices)
		for resp := range data {
			prices <- &resp.Data
		}
	}()

	return prices, nil
}

// WatchOrderBook implements provider.OrderBookStreamProvider.
func (c *Client) WatchOrderBook(ctx context.Context, symbol string, market model.MarketType, depth int) (<-chan *model.OrderBook, error) {
	pool, out, err := c.pool(ctx)
	if err != nil {
		return nil, err
	}

	instType := "SPOT"
	if market == model.MarketFutures {
		instType = "USDT-FUTURES"
	}

	channel := "books5"
	if depth > 0 {
		if depth <= 5 {
			channel = "books5"
		} else if depth <= 15 {
			channel = "books15"
		} else {
			channel = "books"
		}
	}

	if err := pool.Subscribe(ctx, ws.Subscription{
		Key: channel + ":" + symbol + ":" + instType,
		Params: []map[string]string{
			{
				"instType": instType,
				"channel":  channel,
				"instId":   symbol,
			},
		},
	}); err != nil {
		pool.Stop()
		return nil, err
	}

	// Extract typed order books
	data, errs := ws.TypedChan[model.OrderBook](out, 100)

	// Drain error channel in background
	go func() {
		for err := range errs {
			logger.Default.Warn("bitget orderbook stream error", "err", err)
		}
	}()

	// Unwrap Response[OrderBook] to raw OrderBook and set market
	books := make(chan *model.OrderBook, 100)
	go func() {
		defer close(books)
		for resp := range data {
			resp.Data.Market = market
			books <- &resp.Data
		}
	}()

	return books, nil
}
