package mexc

import (
	"context"
	"strings"
	"time"

	"github.com/mdnmdn/bits/internal/logger"
	"github.com/mdnmdn/bits/internal/ws"
	"github.com/mdnmdn/bits/model"
	"github.com/mdnmdn/bits/provider"
)

func (c *Client) ensurePriceStream(ctx context.Context, market model.MarketType) error {
	c.streamMu.Lock()
	defer c.streamMu.Unlock()

	logger.Default.Debug("mexc: ensurePriceStream called", "market", market)

	isFutures := market == model.MarketFutures
	url := spotWSURL
	if isFutures {
		url = futuresWSURL
	}

	// Check if we already have a running pool for this market
	if c.pricePool != nil && c.priceStatus == provider.StreamStatusRunning && c.priceMarket == market {
		logger.Default.Debug("mexc: price pool already running for market", "market", market)
		return nil
	}

	// Stop existing pool if switching market
	if c.pricePool != nil {
		logger.Default.Debug("mexc: stopping existing price pool")
		c.pricePool.Stop()
	}

	logger.Default.Debug("mexc: creating new pool", "url", url, "isFutures", isFutures)

	cfg := ws.SessionConfig{
		ConnConfig: ws.ConnConfig{
			URL:           url,
			PingInterval:  60 * time.Second,
			OutChanBuffer: 100,
		},
		Protocol: newMEXCProtocol(providerID, isFutures),
	}
	pool := ws.NewPool(cfg, 0)
	out, err := pool.Start(ctx)
	if err != nil {
		logger.Default.Error("mexc: failed to start pool", "err", err)
		return err
	}

	logger.Default.Debug("mexc: pool started successfully")

	c.pricePool = pool
	c.priceOut = out
	c.priceStatus = provider.StreamStatusRunning
	c.priceMarket = market

	go c.runPriceDispatcher(ctx, market)

	return nil
}

func (c *Client) runPriceDispatcher(ctx context.Context, market model.MarketType) {
	data, errs := ws.TypedChan[model.CoinPrice](c.priceOut, 100)

	go func() {
		for err := range errs {
			logger.Default.Warn("mexc price stream error", "err", err)
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
	return c.SubscribePrice(ctx, ids)
}

func (c *Client) SubscribePrice(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error) {
	// Default to spot market
	market := model.MarketSpot

	// Check if any ID contains underscore (futures symbol format)
	for _, id := range ids {
		if strings.Contains(id, "_") {
			market = model.MarketFutures
			break
		}
	}

	if err := c.ensurePriceStream(ctx, market); err != nil {
		return nil, err
	}

	c.streamMu.Lock()
	defer c.streamMu.Unlock()

	for _, id := range ids {
		if _, exists := c.priceSubs[id]; !exists {
			c.priceSubs[id] = true

			var topic string
			if market == model.MarketFutures {
				topic = "ticker|" + id
			} else {
				topic = "spot@public.miniTicker.v3.api.pb@" + id + "@UTC+8"
			}

			logger.Default.Debug("mexc: subscribing to price", "id", id, "topic", topic)

			_ = c.pricePool.Subscribe(ctx, ws.Subscription{
				Key:    "ticker:" + id + ":" + string(market),
				Params: topic,
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
				market := c.priceMarket
				if market == model.MarketFutures {
					_ = c.pricePool.Unsubscribe(ctx, "ticker:"+id+":futures")
				} else {
					_ = c.pricePool.Unsubscribe(ctx, "ticker:"+id+":spot")
				}
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
	market := c.priceMarket
	subs := make(map[string]bool)
	for k, v := range c.priceSubs {
		subs[k] = v
	}
	c.streamMu.Unlock()

	if c.pricePool != nil {
		c.pricePool.Stop()
	}

	isFutures := market == model.MarketFutures
	url := spotWSURL
	if isFutures {
		url = futuresWSURL
	}

	cfg := ws.SessionConfig{
		ConnConfig: ws.ConnConfig{
			URL:           url,
			PingInterval:  60 * time.Second,
			OutChanBuffer: 100,
		},
		Protocol: newMEXCProtocol(providerID, isFutures),
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
		var topic string
		if market == model.MarketFutures {
			topic = "ticker|" + id
		} else {
			topic = "spot@public.miniTicker.v3.api.pb@" + id + "@UTC+8"
		}
		_ = c.pricePool.Subscribe(ctx, ws.Subscription{
			Key:    "ticker:" + id + ":" + string(market),
			Params: topic,
		})
	}

	go c.runPriceDispatcher(ctx, market)

	return nil
}

func (c *Client) GetDataChannelPrice() <-chan *model.CoinPrice {
	c.streamMu.RLock()
	defer c.streamMu.RUnlock()
	return c.priceChan
}

// Order Book Stream

func (c *Client) ensureBookStream(ctx context.Context, market model.MarketType, depth int) error {
	c.streamMu.Lock()
	defer c.streamMu.Unlock()

	isFutures := market == model.MarketFutures
	url := spotWSURL
	if isFutures {
		url = futuresWSURL
	}

	if c.bookPool != nil && c.bookStatus == provider.StreamStatusRunning &&
		c.bookMarket == market && c.bookDepth == depth {
		return nil
	}

	if c.bookPool != nil {
		c.bookPool.Stop()
	}

	cfg := ws.SessionConfig{
		ConnConfig: ws.ConnConfig{
			URL:           url,
			PingInterval:  60 * time.Second,
			OutChanBuffer: 100,
		},
		Protocol: newMEXCProtocol(providerID, isFutures),
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
			logger.Default.Warn("mexc orderbook stream error", "err", err)
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
	return c.SubscribeOrderBook(ctx, symbols, market, depth)
}

func (c *Client) SubscribeOrderBook(ctx context.Context, symbols []string, market model.MarketType, depth int) (<-chan *model.OrderBook, error) {
	if err := c.ensureBookStream(ctx, market, depth); err != nil {
		return nil, err
	}

	c.streamMu.Lock()
	defer c.streamMu.Unlock()

	var channelPrefix string
	if market == model.MarketFutures {
		channelPrefix = "depth"
	} else {
		level := 5
		if depth > 0 {
			if depth <= 5 {
				level = 5
			} else if depth <= 10 {
				level = 10
			} else {
				level = 20
			}
		}
		channelPrefix = "limit.depth.v3.api.pb"
		_ = level // unused for now
	}

	for _, symbol := range symbols {
		if _, exists := c.bookSubs[symbol]; !exists {
			c.bookSubs[symbol] = true

			var topic string
			if market == model.MarketFutures {
				topic = channelPrefix + "|" + symbol
			} else {
				// Format: spot@public.limit.depth.v3.api.pb@<symbol>@<level>
				topic = "spot@public." + channelPrefix + "@" + symbol + "@5"
			}

			_ = c.bookPool.Subscribe(ctx, ws.Subscription{
				Key:    channelPrefix + ":" + symbol + ":" + string(market),
				Params: topic,
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
				channel := "limit.depth.v3.api.pb@5"
				if c.bookMarket == model.MarketFutures {
					channel = "depth"
				}
				_ = c.bookPool.Unsubscribe(ctx, channel+":"+symbol+":"+string(c.bookMarket))
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

	isFutures := market == model.MarketFutures
	url := spotWSURL
	if isFutures {
		url = futuresWSURL
	}

	cfg := ws.SessionConfig{
		ConnConfig: ws.ConnConfig{
			URL:           url,
			PingInterval:  60 * time.Second,
			OutChanBuffer: 100,
		},
		Protocol: newMEXCProtocol(providerID, isFutures),
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

	var channel string
	if market == model.MarketFutures {
		channel = "depth"
	} else {
		level := "5"
		if depth > 0 {
			if depth <= 5 {
				level = "5"
			} else if depth <= 10 {
				level = "10"
			} else {
				level = "20"
			}
		}
		channel = "limit.depth.v3.api.pb@" + level
	}

	for symbol := range subs {
		var topic string
		if market == model.MarketFutures {
			topic = channel + "|" + symbol
		} else {
			topic = "spot@public." + channel + "@" + symbol
		}
		_ = c.bookPool.Subscribe(ctx, ws.Subscription{
			Key:    channel + ":" + symbol + ":" + string(market),
			Params: topic,
		})
	}

	go c.runBookDispatcher(ctx, market)

	return nil
}

func (c *Client) GetDataChannelOrderBook() <-chan *model.OrderBook {
	c.streamMu.RLock()
	defer c.streamMu.RUnlock()
	return c.bookChan
}
