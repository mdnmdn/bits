package bitget

import (
	"context"
	"time"

	"github.com/mdnmdn/bits/internal/logger"
	"github.com/mdnmdn/bits/internal/ws"
	"github.com/mdnmdn/bits/internal/ws/middleware"
	"github.com/mdnmdn/bits/model"
	"github.com/mdnmdn/bits/provider"
)

const wsURL = "wss://ws.bitget.com/v2/ws/public"

func (c *Client) ensurePriceStream(ctx context.Context) error {
	c.streamMu.Lock()
	defer c.streamMu.Unlock()

	if c.pricePool != nil && c.priceStatus == provider.StreamStatusRunning {
		return nil
	}

	cfg := ws.SessionConfig{
		ConnConfig: ws.ConnConfig{
			URL:           wsURL,
			PingInterval:  30 * time.Second,
			OutChanBuffer: 100,
		},
		Protocol: &bitgetProtocol{providerID: providerID},
	}
	pool := ws.NewPool(cfg, 0)
	out, err := pool.Start(ctx)
	if err != nil {
		return err
	}

	c.pricePool = pool
	c.priceOut = out
	c.priceStatus = provider.StreamStatusRunning

	go c.runPriceDispatcher(ctx)

	return nil
}

func (c *Client) runPriceDispatcher(ctx context.Context) {
	data, errs := ws.TypedChan[model.CoinPrice](c.priceOut, 100)

	go func() {
		for err := range errs {
			logger.Default.Warn("bitget price stream error", "err", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			c.streamMu.Lock()
			c.priceStatus = provider.StreamStatusStopped
			if c.priceChan != nil {
				close(c.priceChan)
				c.priceChan = nil
			}
			c.streamMu.Unlock()
			return
		case resp, ok := <-data:
			if !ok {
				c.streamMu.Lock()
				c.priceStatus = provider.StreamStatusStopped
				if c.priceChan != nil {
					close(c.priceChan)
					c.priceChan = nil
				}
				c.streamMu.Unlock()
				return
			}

			price := &resp.Data
			c.streamMu.RLock()
			ch := c.priceChan
			c.streamMu.RUnlock()

			if ch != nil {
				select {
				case ch <- price:
				default:
				}
			}
		}
	}
}

func (c *Client) StartPriceStream(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error) {
	if err := c.ensurePriceStream(ctx); err != nil {
		return nil, err
	}

	c.streamMu.Lock()
	defer c.streamMu.Unlock()

	for _, id := range ids {
		if _, exists := c.priceSubs[id]; !exists {
			c.priceSubs[id] = true

			_ = c.pricePool.Subscribe(ctx, ws.Subscription{
				Key: "ticker:" + id + ":SPOT",
				Params: []map[string]string{
					{
						"instType": "SPOT",
						"channel":  "ticker",
						"instId":   id,
					},
				},
			})
		}
	}

	return c.priceChan, nil
}

func (c *Client) SubscribePrice(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error) {
	if err := c.ensurePriceStream(ctx); err != nil {
		return nil, err
	}

	c.streamMu.Lock()
	defer c.streamMu.Unlock()

	for _, id := range ids {
		if _, exists := c.priceSubs[id]; !exists {
			c.priceSubs[id] = true

			_ = c.pricePool.Subscribe(ctx, ws.Subscription{
				Key: "ticker:" + id + ":SPOT",
				Params: []map[string]string{
					{
						"instType": "SPOT",
						"channel":  "ticker",
						"instId":   id,
					},
				},
			})
		}
	}

	return c.priceChan, nil
}

func (c *Client) UnsubscribePrice(ctx context.Context, ids []string) error {
	c.streamMu.Lock()
	defer c.streamMu.Unlock()

	for _, id := range ids {
		if _, exists := c.priceSubs[id]; exists {
			delete(c.priceSubs, id)

			if c.pricePool != nil {
				_ = c.pricePool.Unsubscribe(ctx, "ticker:"+id+":SPOT")
			}
		}
	}

	return nil
}

func (c *Client) SubscribedPrices() []string {
	c.streamMu.RLock()
	defer c.streamMu.RUnlock()

	ids := make([]string, 0, len(c.priceSubs))
	for id := range c.priceSubs {
		ids = append(ids, id)
	}
	return ids
}

func (c *Client) StopPriceStream() error {
	c.streamMu.Lock()
	defer c.streamMu.Unlock()

	c.priceSubs = make(map[string]bool)

	if c.priceChan != nil {
		close(c.priceChan)
		c.priceChan = nil
	}

	if c.pricePool != nil {
		c.pricePool.Stop()
		c.pricePool = nil
	}

	c.priceStatus = provider.StreamStatusStopped
	return nil
}

func (c *Client) PriceStreamStatus() provider.StreamStatus {
	c.streamMu.RLock()
	defer c.streamMu.RUnlock()
	return c.priceStatus
}

func (c *Client) GetLastPrice(id string) (*model.CoinPrice, error) {
	return nil, nil
}

func (c *Client) ReconnectPrice(ctx context.Context) error {
	c.streamMu.Lock()
	subs := make(map[string]bool)
	for k, v := range c.priceSubs {
		subs[k] = v
	}
	c.streamMu.Unlock()

	if c.pricePool != nil {
		c.pricePool.Stop()
	}

	cfg := ws.SessionConfig{
		ConnConfig: ws.ConnConfig{
			URL:           wsURL,
			PingInterval:  30 * time.Second,
			OutChanBuffer: 100,
		},
		Protocol: &bitgetProtocol{providerID: providerID},
	}
	pool := ws.NewPool(cfg, 0)
	out, err := pool.Start(ctx)
	if err != nil {
		c.streamMu.Lock()
		c.priceStatus = provider.StreamStatusError
		c.streamMu.Unlock()
		return err
	}

	c.streamMu.Lock()
	c.pricePool = pool
	c.priceOut = out
	c.priceStatus = provider.StreamStatusRunning
	c.priceSubs = subs
	c.streamMu.Unlock()

	for id := range subs {
		_ = c.pricePool.Subscribe(ctx, ws.Subscription{
			Key: "ticker:" + id + ":SPOT",
			Params: []map[string]string{
				{
					"instType": "SPOT",
					"channel":  "ticker",
					"instId":   id,
				},
			},
		})
	}

	go c.runPriceDispatcher(ctx)

	return nil
}

func (c *Client) ensureBookStream(ctx context.Context, market model.MarketType, depth int) error {
	c.streamMu.Lock()
	defer c.streamMu.Unlock()

	if c.bookPool != nil && c.bookStatus == provider.StreamStatusRunning &&
		c.bookMarket == market && c.bookDepth == depth {
		return nil
	}

	if c.bookPool != nil {
		c.bookPool.Stop()
	}

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
	pool := ws.NewPool(cfg, 0)
	out, err := pool.Start(ctx)
	if err != nil {
		return err
	}

	c.bookPool = pool
	c.bookOut = out
	c.bookStatus = provider.StreamStatusRunning
	c.bookMarket = market
	c.bookDepth = depth

	go c.runBookDispatcher(ctx, market)

	return nil
}

func (c *Client) runBookDispatcher(ctx context.Context, market model.MarketType) {
	data, errs := ws.TypedChan[model.OrderBook](c.bookOut, 100)

	go func() {
		for err := range errs {
			logger.Default.Warn("bitget orderbook stream error", "err", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			c.streamMu.Lock()
			c.bookStatus = provider.StreamStatusStopped
			if c.bookChan != nil {
				close(c.bookChan)
				c.bookChan = nil
			}
			c.streamMu.Unlock()
			return
		case resp, ok := <-data:
			if !ok {
				c.streamMu.Lock()
				c.bookStatus = provider.StreamStatusStopped
				if c.bookChan != nil {
					close(c.bookChan)
					c.bookChan = nil
				}
				c.streamMu.Unlock()
				return
			}

			book := &resp.Data
			book.Market = market

			c.streamMu.RLock()
			ch := c.bookChan
			c.streamMu.RUnlock()

			if ch != nil {
				select {
				case ch <- book:
				default:
				}
			}
		}
	}
}

func (c *Client) StartOrderBookStream(ctx context.Context, symbols []string, market model.MarketType, depth int) (<-chan *model.OrderBook, error) {
	if err := c.ensureBookStream(ctx, market, depth); err != nil {
		return nil, err
	}

	c.streamMu.Lock()
	defer c.streamMu.Unlock()

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

	for _, symbol := range symbols {
		if _, exists := c.bookSubs[symbol]; !exists {
			c.bookSubs[symbol] = true

			_ = c.bookPool.Subscribe(ctx, ws.Subscription{
				Key: channel + ":" + symbol + ":" + instType,
				Params: []map[string]string{
					{
						"instType": instType,
						"channel":  channel,
						"instId":   symbol,
					},
				},
			})
		}
	}

	return c.bookChan, nil
}

func (c *Client) SubscribeOrderBook(ctx context.Context, symbols []string, market model.MarketType, depth int) (<-chan *model.OrderBook, error) {
	if err := c.ensureBookStream(ctx, market, depth); err != nil {
		return nil, err
	}

	c.streamMu.Lock()
	defer c.streamMu.Unlock()

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

	for _, symbol := range symbols {
		if _, exists := c.bookSubs[symbol]; !exists {
			c.bookSubs[symbol] = true

			_ = c.bookPool.Subscribe(ctx, ws.Subscription{
				Key: channel + ":" + symbol + ":" + instType,
				Params: []map[string]string{
					{
						"instType": instType,
						"channel":  channel,
						"instId":   symbol,
					},
				},
			})
		}
	}

	return c.bookChan, nil
}

func (c *Client) UnsubscribeOrderBook(ctx context.Context, symbols []string) error {
	c.streamMu.Lock()
	defer c.streamMu.Unlock()

	for _, symbol := range symbols {
		if _, exists := c.bookSubs[symbol]; exists {
			delete(c.bookSubs, symbol)

			if c.bookPool != nil {
				_ = c.bookPool.Unsubscribe(ctx, "books5:"+symbol+":SPOT")
			}
		}
	}

	return nil
}

func (c *Client) SubscribedOrderBooks() []string {
	c.streamMu.RLock()
	defer c.streamMu.RUnlock()

	symbols := make([]string, 0, len(c.bookSubs))
	for symbol := range c.bookSubs {
		symbols = append(symbols, symbol)
	}
	return symbols
}

func (c *Client) StopOrderBookStream() error {
	c.streamMu.Lock()
	defer c.streamMu.Unlock()

	c.bookSubs = make(map[string]bool)

	if c.bookChan != nil {
		close(c.bookChan)
		c.bookChan = nil
	}

	if c.bookPool != nil {
		c.bookPool.Stop()
		c.bookPool = nil
	}

	c.bookStatus = provider.StreamStatusStopped
	return nil
}

func (c *Client) OrderBookStreamStatus() provider.StreamStatus {
	c.streamMu.RLock()
	defer c.streamMu.RUnlock()
	return c.bookStatus
}

func (c *Client) GetLastOrderBook(symbol string) (*model.OrderBook, error) {
	return nil, nil
}

func (c *Client) ReconnectOrderBook(ctx context.Context) error {
	c.streamMu.Lock()
	market := c.bookMarket
	depth := c.bookDepth
	subs := make(map[string]bool)
	for k, v := range c.bookSubs {
		subs[k] = v
	}
	c.streamMu.Unlock()

	if c.bookPool != nil {
		c.bookPool.Stop()
	}

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
	pool := ws.NewPool(cfg, 0)
	out, err := pool.Start(ctx)
	if err != nil {
		c.streamMu.Lock()
		c.bookStatus = provider.StreamStatusError
		c.streamMu.Unlock()
		return err
	}

	c.streamMu.Lock()
	c.bookPool = pool
	c.bookOut = out
	c.bookStatus = provider.StreamStatusRunning
	c.bookSubs = subs
	c.streamMu.Unlock()

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

	for symbol := range subs {
		_ = c.bookPool.Subscribe(ctx, ws.Subscription{
			Key: channel + ":" + symbol + ":" + instType,
			Params: []map[string]string{
				{
					"instType": instType,
					"channel":  channel,
					"instId":   symbol,
				},
			},
		})
	}

	go c.runBookDispatcher(ctx, market)

	return nil
}

func (c *Client) GetDataChannelPrice() <-chan *model.CoinPrice {
	c.streamMu.RLock()
	defer c.streamMu.RUnlock()
	return c.priceChan
}

func (c *Client) GetDataChannelOrderBook() <-chan *model.OrderBook {
	c.streamMu.RLock()
	defer c.streamMu.RUnlock()
	return c.bookChan
}
