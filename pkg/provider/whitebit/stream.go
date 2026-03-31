package whitebit

import (
	"context"
	"sync"
	"time"

	"github.com/mdnmdn/bits/internal/logger"
	"github.com/mdnmdn/bits/internal/ws"
	"github.com/mdnmdn/bits/internal/ws/middleware"
	"github.com/mdnmdn/bits/pkg/model"
	"github.com/mdnmdn/bits/pkg/provider"
)

// Dual-Subscription Price Stream Implementation
// ==============================================
// WhiteBit provides price data through two separate WebSocket streams:
//   1. market_subscribe -> market_update: Provides 24h stats (price, volume, high, low, change)
//   2. bookTicker_subscribe -> bookTicker_update: Provides best bid/ask price and size
//
// To get a complete CoinPrice with ALL fields (price, volume, change, bid, ask, size),
// we subscribe to BOTH streams and merge the data locally.
//
// Flow:
//   - StartPriceStream subscribes to BOTH market_subscribe AND bookTicker_subscribe
//   - runPriceDispatcher receives updates from both sources
//   - For each symbol, we maintain a merged state: marketData + bookTickerData
//   - When either source updates, we emit the merged CoinPrice
//
// This approach works for both Spot and Futures markets.

type mergedPriceState struct {
	mu         sync.RWMutex
	market     *model.CoinPrice // price, volume, change, high, low
	bookTicker *model.CoinPrice // bid, ask, sizes
}

func newMergedPriceState() *mergedPriceState {
	return &mergedPriceState{
		market:     &model.CoinPrice{},
		bookTicker: &model.CoinPrice{},
	}
}

func (m *mergedPriceState) updateMarket(price *model.CoinPrice) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.market = price
}

func (m *mergedPriceState) updateBookTicker(price *model.CoinPrice) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bookTicker = price
}

func (m *mergedPriceState) getMerged() *model.CoinPrice {
	m.mu.RLock()
	defer m.mu.RUnlock()

	merged := *m.market

	// Merge book ticker data (bid/ask)
	if m.bookTicker.BidPrice != nil {
		merged.BidPrice = m.bookTicker.BidPrice
		merged.BidSize = m.bookTicker.BidSize
	}
	if m.bookTicker.AskPrice != nil {
		merged.AskPrice = m.bookTicker.AskPrice
		merged.AskSize = m.bookTicker.AskSize
	}

	// Use the more recent timestamp
	if m.bookTicker.Time != nil {
		merged.Time = m.bookTicker.Time
	}

	return &merged
}

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
		Protocol: &whitebitProtocol{providerID: providerID},
	}
	pool := ws.NewPool(cfg, 0)
	out, err := pool.Start(ctx)
	if err != nil {
		return err
	}

	c.pricePool = pool
	c.priceOut = out
	c.priceStatus = provider.StreamStatusRunning
	c.priceMerged = make(map[string]*mergedPriceState)

	go c.runPriceDispatcher(ctx)

	return nil
}

func (c *Client) runPriceDispatcher(ctx context.Context) {
	data, errs := ws.TypedChan[model.CoinPrice](c.priceOut, 100)

	go func() {
		for err := range errs {
			logger.Default.Warn("whitebit price stream error", "err", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			c.streamMu.Lock()
			c.priceStatus = provider.StreamStatusStopped
			c.priceMerged = nil
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
				c.priceMerged = nil
				if c.priceChan != nil {
					close(c.priceChan)
					c.priceChan = nil
				}
				c.streamMu.Unlock()
				return
			}

			// Check for nil response
			if resp == nil {
				continue
			}

			price := &resp.Data
			if price == nil || price.Symbol == "" {
				continue
			}
			symbol := price.Symbol

			// Determine the source of this update and update the appropriate state
			// We use the presence of certain fields to identify the source:
			// - market_update has: Price, Volume24h, Change24h, High24h, Low24h
			// - bookTicker_update has: BidPrice, AskPrice (but no Price/Volume)
			c.streamMu.Lock()

			// Check if stream was stopped
			if c.priceMerged == nil {
				c.streamMu.Unlock()
				return
			}

			state, exists := c.priceMerged[symbol]
			if !exists {
				state = newMergedPriceState()
				c.priceMerged[symbol] = state
			}

			// Check if this is a market update (has Price field from market_subscribe)
			// or a bookTicker update (has BidPrice but no Volume24h from bookTicker_subscribe)
			if price.Price > 0 && price.Volume24h != nil {
				state.updateMarket(price)
			} else if price.BidPrice != nil && price.AskPrice != nil {
				state.updateBookTicker(price)
			}

			// Get merged result and send to channel
			merged := state.getMerged()

			ch := c.priceChan
			c.streamMu.Unlock()

			if ch != nil {
				select {
				case ch <- merged:
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

			// Subscribe to BOTH market_subscribe AND bookTicker_subscribe
			// market_subscribe provides: price, volume, change, high, low
			// bookTicker_subscribe provides: bid price, bid size, ask price, ask size
			_ = c.pricePool.Subscribe(ctx, ws.Subscription{
				Key: "market:" + id,
				Params: whitebitSubParams{
					Method: "market_subscribe",
					Args:   []any{id},
				},
			})
			_ = c.pricePool.Subscribe(ctx, ws.Subscription{
				Key: "bookTicker:" + id,
				Params: whitebitSubParams{
					Method: "bookTicker_subscribe",
					Args:   []any{id},
				},
			})
		}
	}

	return c.priceChan, nil
}

func (c *Client) SubscribePrice(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error) {
	return c.StartPriceStream(ctx, ids)
}

func (c *Client) UnsubscribePrice(ctx context.Context, ids []string) error {
	c.streamMu.Lock()
	defer c.streamMu.Unlock()

	for _, id := range ids {
		if c.priceSubs[id] {
			delete(c.priceSubs, id)

			// Unsubscribe from both streams
			_ = c.pricePool.Unsubscribe(ctx, "market:"+id)
			_ = c.pricePool.Unsubscribe(ctx, "bookTicker:"+id)
		}
	}
	return nil
}

func (c *Client) SubscribedPrices() []string {
	c.streamMu.RLock()
	defer c.streamMu.RUnlock()

	symbols := make([]string, 0, len(c.priceSubs))
	for s := range c.priceSubs {
		symbols = append(symbols, s)
	}
	return symbols
}

func (c *Client) StopPriceStream() error {
	c.streamMu.Lock()
	defer c.streamMu.Unlock()

	c.priceSubs = make(map[string]bool)

	if c.priceChan != nil {
		close(c.priceChan)
		c.priceChan = nil
	}

	c.priceMerged = nil

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
	c.streamMu.RLock()
	state := c.priceMerged[id]
	ch := c.priceChan
	c.streamMu.RUnlock()

	if state == nil || ch == nil {
		return nil, nil
	}

	return state.getMerged(), nil
}

func (c *Client) ReconnectPrice(ctx context.Context) error {
	c.streamMu.Lock()
	defer c.streamMu.Unlock()

	if c.pricePool != nil {
		c.pricePool.Stop()
	}

	cfg := ws.SessionConfig{
		ConnConfig: ws.ConnConfig{
			URL:           wsURL,
			PingInterval:  30 * time.Second,
			OutChanBuffer: 100,
		},
		Protocol: &whitebitProtocol{providerID: providerID},
	}
	pool := ws.NewPool(cfg, 0)
	out, err := pool.Start(ctx)
	if err != nil {
		return err
	}

	c.pricePool = pool
	c.priceOut = out
	c.priceStatus = provider.StreamStatusRunning
	c.priceMerged = make(map[string]*mergedPriceState)

	subs := make(map[string]bool)
	for s := range c.priceSubs {
		subs[s] = true
	}
	c.priceSubs = make(map[string]bool)

	go c.runPriceDispatcher(ctx)

	for s := range subs {
		c.priceSubs[s] = true
		_ = c.pricePool.Subscribe(ctx, ws.Subscription{
			Key: "market:" + s,
			Params: whitebitSubParams{
				Method: "market_subscribe",
				Args:   []any{s},
			},
		})
		_ = c.pricePool.Subscribe(ctx, ws.Subscription{
			Key: "bookTicker:" + s,
			Params: whitebitSubParams{
				Method: "bookTicker_subscribe",
				Args:   []any{s},
			},
		})
	}

	return nil
}

func (c *Client) GetDataChannelPrice() <-chan *model.CoinPrice {
	c.streamMu.RLock()
	defer c.streamMu.RUnlock()
	return c.priceChan
}

func (c *Client) ensureBookStream(ctx context.Context, market model.MarketType, depth int) error {
	c.streamMu.Lock()
	defer c.streamMu.Unlock()

	if c.bookPool != nil && c.bookStatus == provider.StreamStatusRunning && c.market == market {
		return nil
	}

	c.market = market

	cfg := ws.SessionConfig{
		ConnConfig: ws.ConnConfig{
			URL:           wsURL,
			PingInterval:  30 * time.Second,
			OutChanBuffer: 100,
		},
		Protocol: &whitebitProtocol{providerID: providerID},
		Pipeline: ws.Pipeline{
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

	go c.runBookDispatcher(ctx, market)

	return nil
}

func (c *Client) runBookDispatcher(ctx context.Context, market model.MarketType) {
	data, errs := ws.TypedChan[model.OrderBook](c.bookOut, 100)

	go func() {
		for err := range errs {
			logger.Default.Warn("whitebit orderbook stream error", "err", err)
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

	if depth == 0 {
		depth = 20
	}

	for _, symbol := range symbols {
		if _, exists := c.bookSubs[symbol]; !exists {
			c.bookSubs[symbol] = true

			_ = c.bookPool.Subscribe(ctx, ws.Subscription{
				Key: "depth:" + symbol,
				Params: whitebitSubParams{
					Method: "depth_subscribe",
					Args:   []any{symbol, depth, "0", true},
				},
			})
		}
	}

	return c.bookChan, nil
}

func (c *Client) SubscribeOrderBook(ctx context.Context, symbols []string, market model.MarketType, depth int) (<-chan *model.OrderBook, error) {
	return c.StartOrderBookStream(ctx, symbols, market, depth)
}

func (c *Client) UnsubscribeOrderBook(ctx context.Context, symbols []string) error {
	c.streamMu.Lock()
	defer c.streamMu.Unlock()

	for _, symbol := range symbols {
		if c.bookSubs[symbol] {
			delete(c.bookSubs, symbol)

			_ = c.bookPool.Unsubscribe(ctx, "depth:"+symbol)
		}
	}
	return nil
}

func (c *Client) SubscribedOrderBooks() []string {
	c.streamMu.RLock()
	defer c.streamMu.RUnlock()

	symbols := make([]string, 0, len(c.bookSubs))
	for s := range c.bookSubs {
		symbols = append(symbols, s)
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
	c.streamMu.RLock()
	ch := c.bookChan
	c.streamMu.RUnlock()

	if ch == nil {
		return nil, nil
	}

	select {
	case book := <-ch:
		return book, nil
	default:
		return nil, nil
	}
}

func (c *Client) ReconnectOrderBook(ctx context.Context) error {
	c.streamMu.Lock()
	defer c.streamMu.Unlock()

	if c.bookPool != nil {
		c.bookPool.Stop()
	}

	cfg := ws.SessionConfig{
		ConnConfig: ws.ConnConfig{
			URL:           wsURL,
			PingInterval:  30 * time.Second,
			OutChanBuffer: 100,
		},
		Protocol: &whitebitProtocol{providerID: providerID},
		Pipeline: ws.Pipeline{
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

	market := c.market
	subs := make(map[string]bool)
	for s := range c.bookSubs {
		subs[s] = true
	}
	c.bookSubs = make(map[string]bool)

	go c.runBookDispatcher(ctx, market)

	for s := range subs {
		c.bookSubs[s] = true
		_ = c.bookPool.Subscribe(ctx, ws.Subscription{
			Key: "depth:" + s,
			Params: whitebitSubParams{
				Method: "depth_subscribe",
				Args:   []any{s, 20, "0", true},
			},
		})
	}

	return nil
}

func (c *Client) GetDataChannelOrderBook() <-chan *model.OrderBook {
	c.streamMu.RLock()
	defer c.streamMu.RUnlock()
	return c.bookChan
}

func (c *Client) GetSymbolForMarket(symbol string, market model.MarketType) string {
	return symbol
}
