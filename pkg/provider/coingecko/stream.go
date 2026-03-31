package coingecko

import (
	"context"

	"github.com/mdnmdn/bits/internal/ws"
	"github.com/mdnmdn/bits/pkg/model"
	"github.com/mdnmdn/bits/pkg/provider"
)

func (c *Client) ensurePriceStream(ctx context.Context) error {
	c.streamMu.Lock()
	defer c.streamMu.Unlock()

	if c.wsClient != nil && c.priceStatus == provider.StreamStatusRunning {
		return nil
	}

	c.wsClient = ws.NewClient(c.cfg, c.priceSubsList)
	c.wsClient.UserAgent = c.UserAgent

	legacyCh, err := c.wsClient.Connect(ctx)
	if err != nil {
		return err
	}

	c.priceChan = make(chan *model.CoinPrice, 100)
	c.priceStatus = provider.StreamStatusRunning

	go c.runPriceDispatcher(ctx, legacyCh)

	return nil
}

func (c *Client) runPriceDispatcher(ctx context.Context, legacyCh <-chan *ws.Update) {
	defer close(c.priceChan)

	for update := range legacyCh {
		change := update.Change24h
		cp := &model.CoinPrice{
			ID:        update.CoinID,
			Symbol:    update.CoinID,
			Currency:  "usd",
			Price:     update.Price,
			Change24h: &change,
		}
		select {
		case c.priceChan <- cp:
		case <-ctx.Done():
			return
		}
	}

	c.streamMu.Lock()
	c.priceStatus = provider.StreamStatusStopped
	c.streamMu.Unlock()
}

func (c *Client) StartPriceStream(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error) {
	c.streamMu.Lock()
	c.priceSubsList = ids
	c.streamMu.Unlock()

	if err := c.ensurePriceStream(ctx); err != nil {
		return nil, err
	}

	return c.priceChan, nil
}

func (c *Client) SubscribePrice(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error) {
	c.streamMu.Lock()
	c.priceSubsList = append(c.priceSubsList, ids...)
	c.streamMu.Unlock()

	if err := c.ensurePriceStream(ctx); err != nil {
		return nil, err
	}

	return c.priceChan, nil
}

func (c *Client) UnsubscribePrice(ctx context.Context, ids []string) error {
	// CoinGecko doesn't support unsubscribe
	return nil
}

func (c *Client) SubscribedPrices() []string {
	c.streamMu.RLock()
	defer c.streamMu.RUnlock()
	return c.priceSubsList
}

func (c *Client) StopPriceStream() error {
	c.streamMu.Lock()
	defer c.streamMu.Unlock()

	if c.wsClient != nil {
		_ = c.wsClient.Close()
		c.wsClient = nil
	}

	if c.priceChan != nil {
		close(c.priceChan)
		c.priceChan = nil
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
	subs := make([]string, len(c.priceSubsList))
	copy(subs, c.priceSubsList)
	c.streamMu.Unlock()

	if c.wsClient != nil {
		_ = c.wsClient.Close()
	}

	c.wsClient = ws.NewClient(c.cfg, subs)
	c.wsClient.UserAgent = c.UserAgent

	legacyCh, err := c.wsClient.Connect(ctx)
	if err != nil {
		c.streamMu.Lock()
		c.priceStatus = provider.StreamStatusError
		c.streamMu.Unlock()
		return err
	}

	c.streamMu.Lock()
	c.priceStatus = provider.StreamStatusRunning
	c.streamMu.Unlock()

	go c.runPriceDispatcher(ctx, legacyCh)

	return nil
}

func (c *Client) GetDataChannelPrice() <-chan *model.CoinPrice {
	c.streamMu.RLock()
	defer c.streamMu.RUnlock()
	return c.priceChan
}
