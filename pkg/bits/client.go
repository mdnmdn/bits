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
// # Symbol Normalization (Optional)
//
// The symbol engine is optional. Without it, you must use provider-specific symbol formats:
//
//	// Without symbol engine - manual symbol format required
//	price, _ := client.GetPrice(ctx, "BTCUSDT", "binance")     // OK
//	price, _ := client.GetPrice(ctx, "BTC_USDT", "whitebit")  // OK
//	price, _ := client.GetPrice(ctx, "BTC-USDT", "binance")   // FAILS
//
// Enable it with WithSymbolEngine() for automatic resolution:
//
//	// With symbol engine - use normalized symbols
//	client := bits.NewClient(cfg, bits.WithSymbolEngine())
//	price, _ := client.GetPriceWithResolution(ctx, "BTC-USDT", "binance", "spot")   // OK
//	price, _ := client.GetPriceWithResolution(ctx, "BTC-USDT", "whitebit", "spot")  // OK
//
// NormalizeSymbol works without the engine:
//
//	normalized := bits.NormalizeSymbol("BTC_USDT") // returns "BTC-USDT"
package bits

import (
	"context"
	"fmt"
	"sync"

	"github.com/mdnmdn/bits/pkg/config"
	"github.com/mdnmdn/bits/pkg/model"
	"github.com/mdnmdn/bits/pkg/provider"
	"github.com/mdnmdn/bits/pkg/provider/registry"
	"github.com/mdnmdn/bits/pkg/resolve/symbol"
	"github.com/mdnmdn/bits/pkg/resolve/symbol/translators"
)

// Client acts as a manager for multiple providers.
// It provides convenient methods for fetching prices and managing symbol resolution.
type Client struct {
	Config       *config.Config
	symbolEngine *symbol.SymbolEngine
}

// Option is a functional option for configuring a Client.
type Option func(*Client)

// WithSymbolEngine enables the symbol resolution engine with disk caching.
// This is recommended for applications that process multiple symbols.
//
// # Cache Behavior
//
// The engine caches exchange symbol lists to disk to reduce API calls:
//   - Default TTL: 5 minutes (configurable via config)
//   - First call: fetches from exchange API (~200-500ms)
//   - Subsequent calls: reads from disk cache (~1-5ms)
//
// Use InvalidateSymbolCache() to force-refresh the cache:
//
//	client.InvalidateSymbolCache("binance", "spot") // refresh single market
//	client.InvalidateAllSymbolCache()            // refresh all markets
func WithSymbolEngine() Option {
	return func(c *Client) {
		c.symbolEngine = symbol.NewSymbolEngine(c.Config)
	}
}

// NewClient creates a new bits Client with the given configuration.
//
// By default, symbol resolution is disabled. Use WithSymbolEngine option
// to enable it:
//
//	client := bits.NewClient(cfg, bits.WithSymbolEngine())
func NewClient(cfg *config.Config, opts ...Option) *Client {
	c := &Client{Config: cfg}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// ResolveSymbol converts a user-friendly symbol to the provider-specific format.
//
// Different exchanges use different symbol formats:
//   - Binance: BTCUSDT
//   - WhiteBit: BTC_USDT
//   - Crypto.com: BTC_USDT
//
// This method normalizes input (e.g., "BTC-USDT", "btc_usdt") to the
// provider's expected format.
//
// Examples:
//
//	symbol, err := client.ResolveSymbol(ctx, "BTC-USDT", "binance", "spot")   // "BTCUSDT"
//	symbol, err := client.ResolveSymbol(ctx, "BTC-USDT", "whitebit", "spot") // "BTC_USDT"
//	symbol, err := client.ResolveSymbol(ctx, "eth-usdt", "binance", "spot")   // "ETHUSDT"
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
//
// This is useful for displaying symbols in a consistent way:
//
//	bits.NormalizeSymbol("BTCUSDT")   // "BTC-USDT"
//	bits.NormalizeSymbol("BTC_USDT")  // "BTC-USDT"
//	bits.NormalizeSymbol("BTC/USDT")  // "BTC-USDT"
func NormalizeSymbol(s string) string {
	return translators.NormalizeSymbol(s)
}

func (c *Client) getSymbolEngine() *symbol.SymbolEngine {
	if c.symbolEngine != nil {
		return c.symbolEngine
	}
	return symbol.NewSymbolEngine(c.Config)
}

// GetPrice retrieves the price for a symbol from a specific provider.
//
// Note: The symbol must be in the provider's native format.
// Use ResolveSymbol first if you have a normalized symbol:
//
//	symbol, _ := client.ResolveSymbol(ctx, "BTC-USDT", "binance", "spot")
//	price, err := client.GetPrice(ctx, symbol, "binance")
func (c *Client) GetPrice(ctx context.Context, symbol string, providerID string) (model.Response[model.CoinPrice], error) {
	p, err := registry.NewProvider(providerID, c.Config)
	if err != nil {
		return model.Response[model.CoinPrice]{}, err
	}

	priceProvider, ok := p.(provider.PriceProvider)
	if !ok {
		return model.Response[model.CoinPrice]{}, fmt.Errorf("provider %s does not support fetching prices", providerID)
	}

	res, err := priceProvider.Price(ctx, []string{symbol}, "")
	if err != nil {
		return model.Response[model.CoinPrice]{}, err
	}

	if len(res.Data) == 0 {
		if len(res.Errors) > 0 {
			return model.Response[model.CoinPrice]{}, res.Errors[0].Err
		}
		return model.Response[model.CoinPrice]{}, fmt.Errorf("no price data returned for %s", symbol)
	}

	return model.Response[model.CoinPrice]{
		Kind:     res.Kind,
		Data:     res.Data[0],
		Provider: res.Provider,
		Market:   res.Market,
		Errors:   res.Errors,
	}, nil
}

// GetPriceWithResolution retrieves the price, automatically resolving the symbol
// to the provider's native format. This is a convenience method that combines
// ResolveSymbol and GetPrice.
//
// Examples:
//
//	price, err := client.GetPriceWithResolution(ctx, "BTC-USDT", "binance", "spot")
//	price, err := client.GetPriceWithResolution(ctx, "ETH-USDT", "whitebit", "spot")
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
//
// Note: Each provider may require a different symbol format. Consider using
// ComparePricesWithResolution for automatic symbol normalization.
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
					Errors:   []model.ItemError{{Symbol: symbol, Err: err}},
				}
				return
			}
			results[index] = res
		}(i, id)
	}

	wg.Wait()
	return results, nil
}

// ComparePricesWithResolution retrieves prices from multiple providers,
// automatically resolving symbols to each provider's native format.
//
// Example:
//
//	// Compare BTC price across exchanges with automatic symbol resolution
//	results, err := client.ComparePricesWithResolution(ctx, "BTC-USDT",
//		[]string{"binance", "bitget", "whitebit"}, "spot")
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
					Errors:   []model.ItemError{{Symbol: normalizedSymbol, Err: err}},
				}
				return
			}
			results[index] = res
		}(i, id)
	}

	wg.Wait()
	return results, nil
}
