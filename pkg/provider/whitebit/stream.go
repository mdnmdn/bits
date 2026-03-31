package whitebit

import (
	"context"

	"github.com/mdnmdn/bits/pkg/model"
	"github.com/mdnmdn/bits/pkg/provider"
)

func (c *Client) StartPriceStream(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error) {
	return nil, nil
}

func (c *Client) SubscribePrice(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error) {
	return nil, nil
}

func (c *Client) UnsubscribePrice(ctx context.Context, ids []string) error {
	return nil
}

func (c *Client) SubscribedPrices() []string {
	return nil
}

func (c *Client) StopPriceStream() error {
	return nil
}

func (c *Client) PriceStreamStatus() provider.StreamStatus {
	return provider.StreamStatusStopped
}

func (c *Client) GetLastPrice(id string) (*model.CoinPrice, error) {
	return nil, nil
}

func (c *Client) ReconnectPrice(ctx context.Context) error {
	return nil
}

func (c *Client) GetDataChannelPrice() <-chan *model.CoinPrice {
	return nil
}

func (c *Client) StartOrderBookStream(ctx context.Context, symbols []string, market model.MarketType, depth int) (<-chan *model.OrderBook, error) {
	return nil, nil
}

func (c *Client) SubscribeOrderBook(ctx context.Context, symbols []string, market model.MarketType, depth int) (<-chan *model.OrderBook, error) {
	return nil, nil
}

func (c *Client) UnsubscribeOrderBook(ctx context.Context, symbols []string) error {
	return nil
}

func (c *Client) SubscribedOrderBooks() []string {
	return nil
}

func (c *Client) StopOrderBookStream() error {
	return nil
}

func (c *Client) OrderBookStreamStatus() provider.StreamStatus {
	return provider.StreamStatusStopped
}

func (c *Client) GetLastOrderBook(symbol string) (*model.OrderBook, error) {
	return nil, nil
}

func (c *Client) ReconnectOrderBook(ctx context.Context) error {
	return nil
}

func (c *Client) GetDataChannelOrderBook() <-chan *model.OrderBook {
	return nil
}
