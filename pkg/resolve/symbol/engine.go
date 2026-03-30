// Package symbol provides symbol resolution and normalization.
//
// There are two entry points:
//
// 1. SymbolEngine (engine.go) - Newer, used by pkg/bits library, has disk caching
// 2. SymbolResolver (resolver.go) - Older, used by CLI commands, in-memory only
//
// The SymbolEngine is the recommended entry point for library users.
// SymbolResolver is kept for CLI compatibility.
package symbol

import (
	"context"
	"fmt"
	"sync"

	"github.com/mdnmdn/bits/pkg/config"
	"github.com/mdnmdn/bits/pkg/model"
	"github.com/mdnmdn/bits/pkg/provider"
	"github.com/mdnmdn/bits/pkg/provider/registry"
	"github.com/mdnmdn/bits/pkg/resolve/symbol/translators"
)

type SymbolEngine struct {
	mu          sync.RWMutex
	providers   map[string]provider.Provider
	lookups     map[string]*lookupTable
	translators map[string]translators.SymbolTranslator
	cfg         *config.Config
	cache       *diskCache
}

type lookupKey struct {
	provider string
	market   model.MarketType
}

func NewSymbolEngine(cfg *config.Config) *SymbolEngine {
	symbolCfg := cfg.Symbol
	return &SymbolEngine{
		providers:   make(map[string]provider.Provider),
		lookups:     make(map[string]*lookupTable),
		translators: make(map[string]translators.SymbolTranslator),
		cfg:         cfg,
		cache:       newDiskCache(symbolCfg.GetCacheDir(), symbolCfg.GetCacheTTL()),
	}
}

func (e *SymbolEngine) getProvider(id string) (provider.Provider, error) {
	e.mu.RLock()
	p, ok := e.providers[id]
	e.mu.RUnlock()

	if ok {
		return p, nil
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if p, ok = e.providers[id]; ok {
		return p, nil
	}

	p, err := registry.NewProvider(id, e.cfg)
	if err != nil {
		return nil, err
	}

	e.providers[id] = p
	return p, nil
}

func (e *SymbolEngine) getLookup(ctx context.Context, providerID string, market model.MarketType) (*lookupTable, error) {
	key := lookupKey{provider: providerID, market: market}

	e.mu.RLock()
	lk, ok := e.lookups[keyString(key)]
	e.mu.RUnlock()

	if ok && lk != nil {
		return lk, nil
	}

	p, err := e.getProvider(providerID)
	if err != nil {
		return nil, err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if lk, ok = e.lookups[keyString(key)]; ok && lk != nil {
		return lk, nil
	}

	symbols, err := e.loadSymbolsFromProvider(ctx, p, providerID, market)
	if err != nil {
		return nil, err
	}

	lk = newLookupTable(symbols)
	e.lookups[keyString(key)] = lk

	return lk, nil
}

func (e *SymbolEngine) loadSymbolsFromProvider(ctx context.Context, p provider.Provider, providerID string, market model.MarketType) ([]model.Symbol, error) {
	symbols, found, err := e.cache.get(providerID, string(market))
	if err == nil && found {
		return symbols, nil
	}

	ep, ok := p.(provider.ExchangeProvider)
	if !ok {
		return nil, fmt.Errorf("provider %s does not support ExchangeInfo", providerID)
	}

	resp, err := ep.ExchangeInfo(ctx, market)
	if err != nil {
		if symbols != nil {
			return symbols, nil
		}
		return nil, err
	}

	_ = e.cache.set(providerID, string(market), resp.Data.Symbols)

	return resp.Data.Symbols, nil
}

func (e *SymbolEngine) Resolve(ctx context.Context, providerID string, input string, market model.MarketType) (string, error) {
	if input == "" {
		return "", nil
	}

	sym, err := e.ResolveToModel(ctx, providerID, input, market)
	if err != nil {
		return input, err
	}

	if sym == nil {
		return input, nil
	}

	return sym.Symbol, nil
}

func (e *SymbolEngine) ResolveToModel(ctx context.Context, providerID string, input string, market model.MarketType) (*model.Symbol, error) {
	if input == "" {
		return nil, nil
	}

	translator := e.getTranslator(providerID)
	base, quote, err := translator.NormalizeInput(input)
	if err != nil || base == "" || quote == "" {
		return nil, nil
	}

	lk, err := e.getLookup(ctx, providerID, market)
	if err != nil {
		return nil, err
	}

	sym := lk.find(base, quote)
	return sym, nil
}

func (e *SymbolEngine) ToNormalized(providerID string, providerSymbol string, market model.MarketType) string {
	translator := e.getTranslator(providerID)
	return translator.ToNormalized(providerSymbol, market)
}

func (e *SymbolEngine) Refresh(ctx context.Context, providerID string, market model.MarketType) error {
	e.cache.invalidate(providerID, string(market))

	e.mu.Lock()
	key := lookupKey{provider: providerID, market: market}
	delete(e.lookups, keyString(key))
	e.mu.Unlock()

	_, err := e.getLookup(ctx, providerID, market)
	return err
}

func (e *SymbolEngine) Invalidate(providerID string, market model.MarketType) {
	e.cache.invalidate(providerID, string(market))

	e.mu.Lock()
	key := lookupKey{provider: providerID, market: market}
	delete(e.lookups, keyString(key))
	e.mu.Unlock()
}

func (e *SymbolEngine) InvalidateAll() {
	e.cache.invalidateAll()

	e.mu.Lock()
	e.lookups = make(map[string]*lookupTable)
	e.mu.Unlock()
}

func (e *SymbolEngine) getTranslator(providerID string) translators.SymbolTranslator {
	e.mu.RLock()
	t, ok := e.translators[providerID]
	e.mu.RUnlock()

	if ok {
		return t
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if t, ok = e.translators[providerID]; ok {
		return t
	}

	t = getTranslator(providerID)
	e.translators[providerID] = t

	return t
}

func keyString(k lookupKey) string {
	return k.provider + "_" + string(k.market)
}

var translatorRegistry = make(map[string]translators.SymbolTranslator)

func getTranslator(providerID string) translators.SymbolTranslator {
	if t, ok := translatorRegistry[providerID]; ok {
		return t
	}
	return translators.NewDefaultTranslator()
}

func init() {
	translatorRegistry["binance"] = translators.NewBinanceTranslator()
	translatorRegistry["whitebit"] = translators.NewWhiteBitTranslator()
	translatorRegistry["cryptocom"] = translators.NewCryptoComTranslator()
	translatorRegistry["mexc"] = translators.NewMEXCTranslator()
	translatorRegistry["bitget"] = translators.NewBitgetTranslator()
}
