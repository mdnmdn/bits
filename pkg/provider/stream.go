package provider

import (
	"context"

	"github.com/mdnmdn/bits/pkg/model"
)

type StreamStatus string

const (
	StreamStatusRunning StreamStatus = "running"
	StreamStatusStopped StreamStatus = "stopped"
	StreamStatusError   StreamStatus = "error"
)

type StreamBook struct {
	Symbol string
	Market model.MarketType
	Book   *model.OrderBook
}

type StreamPrice struct {
	ID    string
	Price *model.CoinPrice
}

// PriceStreamProvider streams live price updates for multiple symbols.
// Uses client as stateful object to manage the WebSocket connection.
type PriceStreamProvider interface {
	// StartPriceStream initiates a price stream for multiple symbols.
	// Returns a single channel where all price updates flow.
	// The ctx controls the lifetime of the stream - cancel to stop.
	StartPriceStream(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error)

	// SubscribePrice adds new symbols to an existing price stream.
	// Returns the single channel where all updates flow.
	SubscribePrice(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error)

	// UnsubscribePrice removes symbols from the price stream.
	UnsubscribePrice(ctx context.Context, ids []string) error

	// SubscribedPrices returns the list of currently subscribed symbol IDs.
	SubscribedPrices() []string

	// StopPriceStream stops all price streams and closes all channels.
	StopPriceStream() error

	// PriceStreamStatus returns the current status of the price stream.
	PriceStreamStatus() StreamStatus

	// GetLastPrice returns the last received price for a symbol, if available.
	GetLastPrice(id string) (*model.CoinPrice, error)

	// ReconnectPrice reconnects the price stream.
	ReconnectPrice(ctx context.Context) error

	// GetDataChannelPrice returns the current price channel.
	// Returns nil if stream is not started.
	GetDataChannelPrice() <-chan *model.CoinPrice
}

// OrderBookStreamProvider streams live order book updates for multiple symbols.
// Uses client as stateful object to manage the WebSocket connection.
type OrderBookStreamProvider interface {
	// StartOrderBookStream initiates an order book stream for multiple symbols.
	// Returns a single channel where all order book updates flow.
	// The ctx controls the lifetime of the stream - cancel to stop.
	StartOrderBookStream(ctx context.Context, symbols []string, market model.MarketType, depth int) (<-chan *model.OrderBook, error)

	// SubscribeOrderBook adds new symbols to an existing order book stream.
	// Returns the single channel where all updates flow.
	SubscribeOrderBook(ctx context.Context, symbols []string, market model.MarketType, depth int) (<-chan *model.OrderBook, error)

	// UnsubscribeOrderBook removes symbols from the order book stream.
	UnsubscribeOrderBook(ctx context.Context, symbols []string) error

	// SubscribedOrderBooks returns the list of currently subscribed symbols.
	SubscribedOrderBooks() []string

	// StopOrderBookStream stops all order book streams and closes all channels.
	StopOrderBookStream() error

	// OrderBookStreamStatus returns the current status of the order book stream.
	OrderBookStreamStatus() StreamStatus

	// GetLastOrderBook returns the last received order book for a symbol, if available.
	GetLastOrderBook(symbol string) (*model.OrderBook, error)

	// ReconnectOrderBook reconnects the order book stream.
	ReconnectOrderBook(ctx context.Context) error

	// GetDataChannelOrderBook returns the current order book channel.
	// Returns nil if stream is not started.
	GetDataChannelOrderBook() <-chan *model.OrderBook
}
