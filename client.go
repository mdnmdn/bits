// Package bits provides a high-level facade for interacting with various crypto providers.
//
// # Getting Started
//
//	cfg := &config.Config{
//		Binance: config.BinanceConfig{
//			Spot: config.MarketConfig{Enabled: true},
//		},
//	}
//	client := bits.NewClient(cfg)
//
// # Provider-Specific Clients
//
// For stateful operations like WebSocket streaming, create a client per provider:
//
//	binance := bits.NewProvider(cfg, "binance")
//	price, err := binance.Price(ctx, []string{"BTCUSDT"}, "")
//
// The client transparently exposes all capability interfaces supported by the provider.
package bits

import (
	"context"
	"sync"

	"github.com/mdnmdn/bits/capability"
	"github.com/mdnmdn/bits/config"
	"github.com/mdnmdn/bits/model"
	"github.com/mdnmdn/bits/provider"
	"github.com/mdnmdn/bits/provider/registry"
	"github.com/mdnmdn/bits/resolve/symbol"
	"github.com/mdnmdn/bits/resolve/symbol/translators"
)

var errNotImplemented = &model.ProviderError{
	Kind:            model.ErrKindUnsupportedFeature,
	ProviderMessage: "not implemented",
}

// Client is a stateful client locked to a specific provider.
// It implements all capability interfaces by delegating to the underlying provider,
// or returning appropriate errors when the provider doesn't support a capability.
type Client struct {
	Config       *config.Config
	provider     provider.Provider
	symbolEngine *symbol.SymbolEngine
}

// Option is a functional option for configuring a Client.
type Option func(*Client)

// WithSymbolEngine enables the symbol resolution engine with disk caching.
func WithSymbolEngine() Option {
	return func(c *Client) {
		c.symbolEngine = symbol.NewSymbolEngine(c.Config)
	}
}

// NewClient creates a new multi-provider Client.
func NewClient(cfg *config.Config, opts ...Option) *Client {
	c := &Client{Config: cfg}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// NewProvider creates a provider-specific client for stateful operations.
// The client is locked to a single provider and exposes all its capabilities.
//
//	cfg := &config.Config{Binance: config.BinanceConfig{Spot: config.MarketConfig{Enabled: true}}}
//	p := bits.NewProvider(cfg, "binance")
func NewProvider(cfg *config.Config, providerID string, opts ...Option) *Client {
	p, err := registry.NewProvider(providerID, cfg)
	if err != nil {
		panic(err)
	}
	c := &Client{Config: cfg, provider: p}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Provider returns the underlying provider instance.
func (c *Client) Provider() provider.Provider {
	return c.provider
}

// ID returns the provider identifier.
func (c *Client) ID() string {
	return c.provider.ID()
}

// Capabilities returns the capability matrix for this provider.
func (c *Client) Capabilities() capability.CapabilityMatrix {
	return c.provider.Capabilities()
}

// ExchangeProvider returns the ExchangeProvider interface if supported.
func (c *Client) ExchangeProvider() provider.ExchangeProvider {
	if p, ok := c.provider.(provider.ExchangeProvider); ok {
		return p
	}
	return &nullExchangeProvider{IDValue: c.ID()}
}

// ServerTime returns the server time from the exchange.
func (c *Client) ServerTime(ctx context.Context) (model.Response[model.ServerTime], error) {
	ep := c.ExchangeProvider()
	if _, ok := ep.(*nullExchangeProvider); ok {
		return model.Response[model.ServerTime]{
			Provider: c.ID(),
			Errors:   []model.ItemError{{Err: errNotImplemented}},
		}, nil
	}
	return ep.ServerTime(ctx)
}

// ExchangeInfo returns the exchange information for a market.
func (c *Client) ExchangeInfo(ctx context.Context, market model.MarketType) (model.Response[model.ExchangeInfo], error) {
	ep := c.ExchangeProvider()
	if _, ok := ep.(*nullExchangeProvider); ok {
		return model.Response[model.ExchangeInfo]{
			Provider: c.ID(),
			Errors:   []model.ItemError{{Err: errNotImplemented}},
		}, nil
	}
	return ep.ExchangeInfo(ctx, market)
}

// AggregatorProvider returns the AggregatorProvider interface if supported.
func (c *Client) AggregatorProvider() provider.AggregatorProvider {
	if p, ok := c.provider.(provider.AggregatorProvider); ok {
		return p
	}
	return &nullAggregatorProvider{IDValue: c.ID()}
}

// CoinMarkets returns ranked coin listings by market cap.
func (c *Client) CoinMarkets(ctx context.Context, opts model.MarketOpts) (model.Response[[]model.CoinMarket], error) {
	ap := c.AggregatorProvider()
	if _, ok := ap.(*nullAggregatorProvider); ok {
		return model.Response[[]model.CoinMarket]{
			Provider: c.ID(),
			Errors:   []model.ItemError{{Err: errNotImplemented}},
		}, nil
	}
	return ap.CoinMarkets(ctx, opts)
}

// PriceProvider returns the PriceProvider interface if supported.
func (c *Client) PriceProvider() provider.PriceProvider {
	if p, ok := c.provider.(provider.PriceProvider); ok {
		return p
	}
	return &nullPriceProvider{IDValue: c.ID()}
}

// Price retrieves the current price for the given IDs or symbols.
func (c *Client) Price(ctx context.Context, ids []string, currency string) (model.Response[[]model.CoinPrice], error) {
	if c.symbolEngine != nil && len(ids) > 0 {
		resolvedIDs := make([]string, len(ids))
		for i, id := range ids {
			resolved := c.resolveSymbolForPrice(ctx, id)
			if resolved == "" {
				resolvedIDs[i] = id
			} else {
				resolvedIDs[i] = resolved
			}
		}
		ids = resolvedIDs
	}
	pp := c.PriceProvider()
	if _, ok := pp.(*nullPriceProvider); ok {
		return model.Response[[]model.CoinPrice]{
			Provider: c.ID(),
			Errors:   []model.ItemError{{Err: errNotImplemented}},
		}, nil
	}
	return pp.Price(ctx, ids, currency)
}

func (c *Client) resolveSymbolForPrice(ctx context.Context, input string) string {
	markets := []model.MarketType{model.MarketSpot, model.MarketFutures, model.MarketMargin}
	for _, market := range markets {
		resolved, err := c.symbolEngine.Resolve(ctx, c.ID(), input, market)
		if err == nil && resolved != "" {
			return resolved
		}
	}
	return ""
}

// CandleProvider returns the CandleProvider interface if supported.
func (c *Client) CandleProvider() provider.CandleProvider {
	if p, ok := c.provider.(provider.CandleProvider); ok {
		return p
	}
	return &nullCandleProvider{IDValue: c.ID()}
}

// Candles retrieves OHLCV candle data.
func (c *Client) Candles(ctx context.Context, symbol string, market model.MarketType, interval string, opts model.CandleOpts) (model.Response[[]model.Candle], error) {
	resolved := c.resolveSymbolIfNeeded(ctx, symbol, market)
	cp := c.CandleProvider()
	if _, ok := cp.(*nullCandleProvider); ok {
		return model.Response[[]model.Candle]{
			Provider: c.ID(),
			Errors:   []model.ItemError{{Err: errNotImplemented}},
		}, nil
	}
	return cp.Candles(ctx, resolved, market, interval, opts)
}

// TickerProvider returns the TickerProvider interface if supported.
func (c *Client) TickerProvider() provider.TickerProvider {
	if p, ok := c.provider.(provider.TickerProvider); ok {
		return p
	}
	return &nullTickerProvider{IDValue: c.ID()}
}

// Ticker24h retrieves 24h rolling ticker statistics.
func (c *Client) Ticker24h(ctx context.Context, symbol string, market model.MarketType) (model.Response[model.Ticker24h], error) {
	resolved := c.resolveSymbolIfNeeded(ctx, symbol, market)
	tp := c.TickerProvider()
	if _, ok := tp.(*nullTickerProvider); ok {
		return model.Response[model.Ticker24h]{
			Provider: c.ID(),
			Errors:   []model.ItemError{{Err: errNotImplemented}},
		}, nil
	}
	return tp.Ticker24h(ctx, resolved, market)
}

// OrderBookProvider returns the OrderBookProvider interface if supported.
func (c *Client) OrderBookProvider() provider.OrderBookProvider {
	if p, ok := c.provider.(provider.OrderBookProvider); ok {
		return p
	}
	return &nullOrderBookProvider{IDValue: c.ID()}
}

// OrderBook retrieves order book depth snapshot.
func (c *Client) OrderBook(ctx context.Context, symbol string, market model.MarketType, depth int) (model.Response[model.OrderBook], error) {
	resolved := c.resolveSymbolIfNeeded(ctx, symbol, market)
	obp := c.OrderBookProvider()
	if _, ok := obp.(*nullOrderBookProvider); ok {
		return model.Response[model.OrderBook]{
			Provider: c.ID(),
			Errors:   []model.ItemError{{Err: errNotImplemented}},
		}, nil
	}
	return obp.OrderBook(ctx, resolved, market, depth)
}

// PriceStreamProvider returns the PriceStreamProvider interface if supported.
func (c *Client) PriceStreamProvider() provider.PriceStreamProvider {
	if p, ok := c.provider.(provider.PriceStreamProvider); ok {
		return p
	}
	return &nullPriceStreamProvider{IDValue: c.ID()}
}

// StartPriceStream initiates a price stream for multiple symbols.
func (c *Client) StartPriceStream(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error) {
	psp := c.PriceStreamProvider()
	if _, ok := psp.(*nullPriceStreamProvider); ok {
		return nil, errNotImplemented
	}
	resolved := make([]string, len(ids))
	for i, id := range ids {
		resolved[i] = c.resolveSymbolIfNeeded(ctx, id, model.MarketSpot)
	}
	return psp.StartPriceStream(ctx, resolved)
}

// SubscribePrice adds new symbols to an existing price stream.
func (c *Client) SubscribePrice(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error) {
	psp := c.PriceStreamProvider()
	if _, ok := psp.(*nullPriceStreamProvider); ok {
		return nil, errNotImplemented
	}
	return psp.SubscribePrice(ctx, ids)
}

// UnsubscribePrice removes symbols from the price stream.
func (c *Client) UnsubscribePrice(ctx context.Context, ids []string) error {
	psp := c.PriceStreamProvider()
	if _, ok := psp.(*nullPriceStreamProvider); ok {
		return errNotImplemented
	}
	return psp.UnsubscribePrice(ctx, ids)
}

// SubscribedPrices returns the list of currently subscribed symbol IDs.
func (c *Client) SubscribedPrices() []string {
	psp := c.PriceStreamProvider()
	if _, ok := psp.(*nullPriceStreamProvider); ok {
		return nil
	}
	return psp.SubscribedPrices()
}

// StopPriceStream stops all price streams and closes all channels.
func (c *Client) StopPriceStream() error {
	psp := c.PriceStreamProvider()
	if _, ok := psp.(*nullPriceStreamProvider); ok {
		return errNotImplemented
	}
	return psp.StopPriceStream()
}

// PriceStreamStatus returns the current status of the price stream.
func (c *Client) PriceStreamStatus() provider.StreamStatus {
	psp := c.PriceStreamProvider()
	if _, ok := psp.(*nullPriceStreamProvider); ok {
		return provider.StreamStatusError
	}
	return psp.PriceStreamStatus()
}

// GetLastPrice returns the last received price for a symbol.
func (c *Client) GetLastPrice(id string) (*model.CoinPrice, error) {
	psp := c.PriceStreamProvider()
	if _, ok := psp.(*nullPriceStreamProvider); ok {
		return nil, errNotImplemented
	}
	return psp.GetLastPrice(id)
}

// ReconnectPrice reconnects the price stream.
func (c *Client) ReconnectPrice(ctx context.Context) error {
	psp := c.PriceStreamProvider()
	if _, ok := psp.(*nullPriceStreamProvider); ok {
		return errNotImplemented
	}
	return psp.ReconnectPrice(ctx)
}

// GetDataChannelPrice returns the current price channel.
func (c *Client) GetDataChannelPrice() <-chan *model.CoinPrice {
	psp := c.PriceStreamProvider()
	if _, ok := psp.(*nullPriceStreamProvider); ok {
		return nil
	}
	return psp.GetDataChannelPrice()
}

// OrderBookStreamProvider returns the OrderBookStreamProvider interface if supported.
func (c *Client) OrderBookStreamProvider() provider.OrderBookStreamProvider {
	if p, ok := c.provider.(provider.OrderBookStreamProvider); ok {
		return p
	}
	return &nullOrderBookStreamProvider{IDValue: c.ID()}
}

// StartOrderBookStream initiates an order book stream for multiple symbols.
func (c *Client) StartOrderBookStream(ctx context.Context, symbols []string, market model.MarketType, depth int) (<-chan *model.OrderBook, error) {
	obsp := c.OrderBookStreamProvider()
	if _, ok := obsp.(*nullOrderBookStreamProvider); ok {
		return nil, errNotImplemented
	}
	resolved := make([]string, len(symbols))
	for i, sym := range symbols {
		resolved[i] = c.resolveSymbolIfNeeded(ctx, sym, market)
	}
	return obsp.StartOrderBookStream(ctx, resolved, market, depth)
}

// SubscribeOrderBook adds new symbols to an existing order book stream.
func (c *Client) SubscribeOrderBook(ctx context.Context, symbols []string, market model.MarketType, depth int) (<-chan *model.OrderBook, error) {
	obsp := c.OrderBookStreamProvider()
	if _, ok := obsp.(*nullOrderBookStreamProvider); ok {
		return nil, errNotImplemented
	}
	return obsp.SubscribeOrderBook(ctx, symbols, market, depth)
}

// UnsubscribeOrderBook removes symbols from the order book stream.
func (c *Client) UnsubscribeOrderBook(ctx context.Context, symbols []string) error {
	obsp := c.OrderBookStreamProvider()
	if _, ok := obsp.(*nullOrderBookStreamProvider); ok {
		return errNotImplemented
	}
	return obsp.UnsubscribeOrderBook(ctx, symbols)
}

// SubscribedOrderBooks returns the list of currently subscribed symbols.
func (c *Client) SubscribedOrderBooks() []string {
	obsp := c.OrderBookStreamProvider()
	if _, ok := obsp.(*nullOrderBookStreamProvider); ok {
		return nil
	}
	return obsp.SubscribedOrderBooks()
}

// StopOrderBookStream stops all order book streams and closes all channels.
func (c *Client) StopOrderBookStream() error {
	obsp := c.OrderBookStreamProvider()
	if _, ok := obsp.(*nullOrderBookStreamProvider); ok {
		return errNotImplemented
	}
	return obsp.StopOrderBookStream()
}

// OrderBookStreamStatus returns the current status of the order book stream.
func (c *Client) OrderBookStreamStatus() provider.StreamStatus {
	obsp := c.OrderBookStreamProvider()
	if _, ok := obsp.(*nullOrderBookStreamProvider); ok {
		return provider.StreamStatusError
	}
	return obsp.OrderBookStreamStatus()
}

// GetLastOrderBook returns the last received order book for a symbol.
func (c *Client) GetLastOrderBook(symbol string) (*model.OrderBook, error) {
	obsp := c.OrderBookStreamProvider()
	if _, ok := obsp.(*nullOrderBookStreamProvider); ok {
		return nil, errNotImplemented
	}
	return obsp.GetLastOrderBook(symbol)
}

// ReconnectOrderBook reconnects the order book stream.
func (c *Client) ReconnectOrderBook(ctx context.Context) error {
	obsp := c.OrderBookStreamProvider()
	if _, ok := obsp.(*nullOrderBookStreamProvider); ok {
		return errNotImplemented
	}
	return obsp.ReconnectOrderBook(ctx)
}

// GetDataChannelOrderBook returns the current order book channel.
func (c *Client) GetDataChannelOrderBook() <-chan *model.OrderBook {
	obsp := c.OrderBookStreamProvider()
	if _, ok := obsp.(*nullOrderBookStreamProvider); ok {
		return nil
	}
	return obsp.GetDataChannelOrderBook()
}

// NullProvider implements all capability interfaces with "not implemented" errors.
// Use it when a provider doesn't support certain capabilities.
type NullProvider struct {
	IDValue string
}

func (n *NullProvider) ID() string             { return n.IDValue }
func (n *NullProvider) SetUserAgent(ua string) {}
func (n *NullProvider) Capabilities() capability.CapabilityMatrix {
	return capability.NewCapabilityMatrix()
}

type nullExchangeProvider struct{ IDValue string }

func (n *nullExchangeProvider) ID() string          { return n.IDValue }
func (n *nullExchangeProvider) SetUserAgent(string) {}
func (n *nullExchangeProvider) Capabilities() capability.CapabilityMatrix {
	return capability.NewCapabilityMatrix()
}
func (n *nullExchangeProvider) ServerTime(ctx context.Context) (model.Response[model.ServerTime], error) {
	return model.Response[model.ServerTime]{Provider: n.ID(), Errors: []model.ItemError{{Err: errNotImplemented}}}, nil
}
func (n *nullExchangeProvider) ExchangeInfo(ctx context.Context, market model.MarketType) (model.Response[model.ExchangeInfo], error) {
	return model.Response[model.ExchangeInfo]{Provider: n.ID(), Errors: []model.ItemError{{Err: errNotImplemented}}}, nil
}

type nullAggregatorProvider struct{ IDValue string }

func (n *nullAggregatorProvider) ID() string          { return n.IDValue }
func (n *nullAggregatorProvider) SetUserAgent(string) {}
func (n *nullAggregatorProvider) Capabilities() capability.CapabilityMatrix {
	return capability.NewCapabilityMatrix()
}
func (n *nullAggregatorProvider) CoinMarkets(ctx context.Context, opts model.MarketOpts) (model.Response[[]model.CoinMarket], error) {
	return model.Response[[]model.CoinMarket]{Provider: n.ID(), Errors: []model.ItemError{{Err: errNotImplemented}}}, nil
}

type nullPriceProvider struct{ IDValue string }

func (n *nullPriceProvider) ID() string          { return n.IDValue }
func (n *nullPriceProvider) SetUserAgent(string) {}
func (n *nullPriceProvider) Capabilities() capability.CapabilityMatrix {
	return capability.NewCapabilityMatrix()
}
func (n *nullPriceProvider) Price(ctx context.Context, ids []string, currency string) (model.Response[[]model.CoinPrice], error) {
	return model.Response[[]model.CoinPrice]{Provider: n.ID(), Errors: []model.ItemError{{Err: errNotImplemented}}}, nil
}

type nullCandleProvider struct{ IDValue string }

func (n *nullCandleProvider) ID() string          { return n.IDValue }
func (n *nullCandleProvider) SetUserAgent(string) {}
func (n *nullCandleProvider) Capabilities() capability.CapabilityMatrix {
	return capability.NewCapabilityMatrix()
}
func (n *nullCandleProvider) Candles(ctx context.Context, symbol string, market model.MarketType, interval string, opts model.CandleOpts) (model.Response[[]model.Candle], error) {
	return model.Response[[]model.Candle]{Provider: n.ID(), Errors: []model.ItemError{{Err: errNotImplemented}}}, nil
}

type nullTickerProvider struct{ IDValue string }

func (n *nullTickerProvider) ID() string          { return n.IDValue }
func (n *nullTickerProvider) SetUserAgent(string) {}
func (n *nullTickerProvider) Capabilities() capability.CapabilityMatrix {
	return capability.NewCapabilityMatrix()
}
func (n *nullTickerProvider) Ticker24h(ctx context.Context, symbol string, market model.MarketType) (model.Response[model.Ticker24h], error) {
	return model.Response[model.Ticker24h]{Provider: n.ID(), Errors: []model.ItemError{{Err: errNotImplemented}}}, nil
}

type nullOrderBookProvider struct{ IDValue string }

func (n *nullOrderBookProvider) ID() string          { return n.IDValue }
func (n *nullOrderBookProvider) SetUserAgent(string) {}
func (n *nullOrderBookProvider) Capabilities() capability.CapabilityMatrix {
	return capability.NewCapabilityMatrix()
}
func (n *nullOrderBookProvider) OrderBook(ctx context.Context, symbol string, market model.MarketType, depth int) (model.Response[model.OrderBook], error) {
	return model.Response[model.OrderBook]{Provider: n.ID(), Errors: []model.ItemError{{Err: errNotImplemented}}}, nil
}

type nullPriceStreamProvider struct{ IDValue string }

func (n *nullPriceStreamProvider) ID() string          { return n.IDValue }
func (n *nullPriceStreamProvider) SetUserAgent(string) {}
func (n *nullPriceStreamProvider) Capabilities() capability.CapabilityMatrix {
	return capability.NewCapabilityMatrix()
}
func (n *nullPriceStreamProvider) StartPriceStream(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error) {
	return nil, errNotImplemented
}
func (n *nullPriceStreamProvider) SubscribePrice(ctx context.Context, ids []string) (<-chan *model.CoinPrice, error) {
	return nil, errNotImplemented
}
func (n *nullPriceStreamProvider) UnsubscribePrice(ctx context.Context, ids []string) error {
	return errNotImplemented
}
func (n *nullPriceStreamProvider) SubscribedPrices() []string { return nil }
func (n *nullPriceStreamProvider) StopPriceStream() error     { return errNotImplemented }
func (n *nullPriceStreamProvider) PriceStreamStatus() provider.StreamStatus {
	return provider.StreamStatusError
}
func (n *nullPriceStreamProvider) GetLastPrice(id string) (*model.CoinPrice, error) {
	return nil, errNotImplemented
}
func (n *nullPriceStreamProvider) ReconnectPrice(ctx context.Context) error     { return errNotImplemented }
func (n *nullPriceStreamProvider) GetDataChannelPrice() <-chan *model.CoinPrice { return nil }

type nullOrderBookStreamProvider struct{ IDValue string }

func (n *nullOrderBookStreamProvider) ID() string          { return n.IDValue }
func (n *nullOrderBookStreamProvider) SetUserAgent(string) {}
func (n *nullOrderBookStreamProvider) Capabilities() capability.CapabilityMatrix {
	return capability.NewCapabilityMatrix()
}
func (n *nullOrderBookStreamProvider) StartOrderBookStream(ctx context.Context, symbols []string, market model.MarketType, depth int) (<-chan *model.OrderBook, error) {
	return nil, errNotImplemented
}
func (n *nullOrderBookStreamProvider) SubscribeOrderBook(ctx context.Context, symbols []string, market model.MarketType, depth int) (<-chan *model.OrderBook, error) {
	return nil, errNotImplemented
}
func (n *nullOrderBookStreamProvider) UnsubscribeOrderBook(ctx context.Context, symbols []string) error {
	return errNotImplemented
}
func (n *nullOrderBookStreamProvider) SubscribedOrderBooks() []string { return nil }
func (n *nullOrderBookStreamProvider) StopOrderBookStream() error     { return errNotImplemented }
func (n *nullOrderBookStreamProvider) OrderBookStreamStatus() provider.StreamStatus {
	return provider.StreamStatusError
}
func (n *nullOrderBookStreamProvider) GetLastOrderBook(symbol string) (*model.OrderBook, error) {
	return nil, errNotImplemented
}
func (n *nullOrderBookStreamProvider) ReconnectOrderBook(ctx context.Context) error {
	return errNotImplemented
}
func (n *nullOrderBookStreamProvider) GetDataChannelOrderBook() <-chan *model.OrderBook { return nil }

// --- Multi-provider Client methods ---

// Provider returns a provider instance by ID (for multi-provider Client).
func (c *Client) GetProvider(providerID string) (provider.Provider, error) {
	return registry.NewProvider(providerID, c.Config)
}

// ResolveSymbol converts a user-friendly symbol to the provider-specific format.
func (c *Client) ResolveSymbol(ctx context.Context, input, providerID string, market model.MarketType) (string, error) {
	engine := c.getSymbolEngine()
	return engine.Resolve(ctx, providerID, input, market)
}

// ResolveSymbolToModel resolves a symbol and returns full symbol metadata.
func (c *Client) ResolveSymbolToModel(ctx context.Context, input, providerID string, market model.MarketType) (*model.Symbol, error) {
	engine := c.getSymbolEngine()
	return engine.ResolveToModel(ctx, providerID, input, market)
}

// InvalidateSymbolCache clears the cached symbol data for a provider and market.
func (c *Client) InvalidateSymbolCache(providerID string, market model.MarketType) {
	if c.symbolEngine != nil {
		c.symbolEngine.Invalidate(providerID, market)
	}
}

// InvalidateAllSymbolCache clears all cached symbol data.
func (c *Client) InvalidateAllSymbolCache() {
	if c.symbolEngine != nil {
		c.symbolEngine.InvalidateAll()
	}
}

// NormalizeSymbol converts any symbol format to a standardized BASE-QUOTE format.
func NormalizeSymbol(s string) string {
	return translators.NormalizeSymbol(s)
}

func (c *Client) getSymbolEngine() *symbol.SymbolEngine {
	if c.symbolEngine != nil {
		return c.symbolEngine
	}
	return symbol.NewSymbolEngine(c.Config)
}

func (c *Client) resolveSymbolIfNeeded(ctx context.Context, symbol string, market model.MarketType) string {
	if c.symbolEngine == nil || symbol == "" {
		return symbol
	}
	resolved, err := c.symbolEngine.Resolve(ctx, c.ID(), symbol, market)
	if err != nil || resolved == "" {
		return symbol
	}
	return resolved
}

// GetPrice retrieves the price for a symbol from a specific provider.
func (c *Client) GetPrice(ctx context.Context, symbol string, providerID string) (model.Response[model.CoinPrice], error) {
	p, err := c.GetProvider(providerID)
	if err != nil {
		return model.Response[model.CoinPrice]{}, err
	}

	priceProvider, ok := p.(provider.PriceProvider)
	if !ok {
		return model.Response[model.CoinPrice]{}, &model.ProviderError{
			Kind:            model.ErrKindUnsupportedFeature,
			ProviderID:      providerID,
			ProviderMessage: "provider does not support fetching prices",
		}
	}

	res, err := priceProvider.Price(ctx, []string{symbol}, "")
	if err != nil {
		return model.Response[model.CoinPrice]{}, err
	}

	if len(res.Data) == 0 {
		if len(res.Errors) > 0 {
			return model.Response[model.CoinPrice]{}, res.Errors[0].Err
		}
		return model.Response[model.CoinPrice]{}, &model.ProviderError{
			Kind:            model.ErrKindNotFound,
			ProviderID:      providerID,
			ProviderMessage: "no price data returned for " + symbol,
		}
	}

	return model.Response[model.CoinPrice]{
		Kind:     res.Kind,
		Data:     res.Data[0],
		Provider: res.Provider,
		Market:   res.Market,
		Errors:   res.Errors,
	}, nil
}

// GetPriceWithResolution retrieves the price, automatically resolving the symbol.
func (c *Client) GetPriceWithResolution(ctx context.Context, normalizedSymbol, providerID string, market model.MarketType) (model.Response[model.CoinPrice], error) {
	resolved, err := c.ResolveSymbol(ctx, normalizedSymbol, providerID, market)
	if err != nil {
		return model.Response[model.CoinPrice]{}, err
	}
	if resolved == "" {
		resolved = normalizedSymbol
	}
	return c.GetPrice(ctx, resolved, providerID)
}

// ComparePrices retrieves the price for a symbol from multiple providers concurrently.
func (c *Client) ComparePrices(ctx context.Context, symbol string, providerIDs []string) ([]model.Response[model.CoinPrice], error) {
	var wg sync.WaitGroup
	results := make([]model.Response[model.CoinPrice], len(providerIDs))

	for i, id := range providerIDs {
		wg.Add(1)
		go func(index int, pid string) {
			defer wg.Done()
			res, err := c.GetPrice(ctx, symbol, pid)
			if err != nil {
				results[index] = model.Response[model.CoinPrice]{
					Provider: pid,
					Errors:   []model.ItemError{{Symbol: symbol, Err: model.WrapError(pid, err)}},
				}
				return
			}
			results[index] = res
		}(i, id)
	}

	wg.Wait()
	return results, nil
}

// ComparePricesWithResolution retrieves prices from multiple providers with automatic symbol resolution.
func (c *Client) ComparePricesWithResolution(ctx context.Context, normalizedSymbol string, providerIDs []string, market model.MarketType) ([]model.Response[model.CoinPrice], error) {
	var wg sync.WaitGroup
	results := make([]model.Response[model.CoinPrice], len(providerIDs))

	for i, id := range providerIDs {
		wg.Add(1)
		go func(index int, pid string) {
			defer wg.Done()
			res, err := c.GetPriceWithResolution(ctx, normalizedSymbol, pid, market)
			if err != nil {
				results[index] = model.Response[model.CoinPrice]{
					Provider: pid,
					Errors:   []model.ItemError{{Symbol: normalizedSymbol, Err: model.WrapError(pid, err)}},
				}
				return
			}
			results[index] = res
		}(i, id)
	}

	wg.Wait()
	return results, nil
}
