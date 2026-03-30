// Package bits provides a high-level facade for interacting with various crypto providers.
package bits

import (
	"context"
	"fmt"
	"sync"

	"github.com/mdnmdn/bits/pkg/config"
	"github.com/mdnmdn/bits/pkg/model"
	"github.com/mdnmdn/bits/pkg/provider"
	"github.com/mdnmdn/bits/pkg/provider/registry"
)

// Client acts as a manager for multiple providers.
type Client struct {
	Config *config.Config
}

// NewClient creates a new bits Client with the given configuration.
func NewClient(cfg *config.Config) *Client {
	return &Client{Config: cfg}
}

// GetPrice retrieves the price for a symbol from a specific provider.
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
