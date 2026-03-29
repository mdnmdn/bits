package symbol

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/mdnmdn/bits/pkg/model"
	"github.com/mdnmdn/bits/pkg/provider"
)

type SymbolResolver struct {
	provider provider.Provider
	symbols  map[model.MarketType][]model.Symbol
	mu       sync.RWMutex
}

func New(p provider.Provider) *SymbolResolver {
	return &SymbolResolver{
		provider: p,
		symbols:  make(map[model.MarketType][]model.Symbol),
	}
}

func (r *SymbolResolver) Resolve(ctx context.Context, input string, market model.MarketType) (string, error) {
	if input == "" {
		return "", nil
	}

	base, quote, err := Normalize(input)
	if err != nil {
		return "", err
	}

	if base == "" || quote == "" {
		return input, nil
	}

	symbols, err := r.loadSymbols(ctx, market)
	if err != nil {
		return input, nil
	}

	for _, sym := range symbols {
		if strings.EqualFold(sym.BaseAsset, base) && strings.EqualFold(sym.QuoteAsset, quote) {
			return sym.Symbol, nil
		}
	}

	return input, nil
}

func (r *SymbolResolver) loadSymbols(ctx context.Context, market model.MarketType) ([]model.Symbol, error) {
	r.mu.RLock()
	if symbols, ok := r.symbols[market]; ok {
		r.mu.RUnlock()
		return symbols, nil
	}
	r.mu.RUnlock()

	ep, ok := r.provider.(provider.ExchangeProvider)
	if !ok {
		return nil, fmt.Errorf("provider does not support ExchangeInfo")
	}

	resp, err := ep.ExchangeInfo(ctx, market)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	r.symbols[market] = resp.Data.Symbols
	r.mu.Unlock()

	return resp.Data.Symbols, nil
}

func (r *SymbolResolver) InvalidateCache(market model.MarketType) {
	r.mu.Lock()
	delete(r.symbols, market)
	r.mu.Unlock()
}

func (r *SymbolResolver) InvalidateAll() {
	r.mu.Lock()
	r.symbols = make(map[model.MarketType][]model.Symbol)
	r.mu.Unlock()
}
