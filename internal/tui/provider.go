package tui

import (
	"context"
	"fmt"

	"github.com/mdnmdn/bits/pkg/capability"
	"github.com/mdnmdn/bits/pkg/config"
	"github.com/mdnmdn/bits/pkg/model"
	"github.com/mdnmdn/bits/pkg/provider"
	"github.com/mdnmdn/bits/pkg/provider/registry"
	"github.com/mdnmdn/bits/pkg/resolve"
)

type TUIProvider struct {
	resolver *resolve.Resolver
	cfg      *config.Config
}

func NewTUIProvider() (*TUIProvider, error) {
	cfg, _, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	return &TUIProvider{
		resolver: resolve.New(cfg, registry.NewProvider, registry.AllProviderIDs),
		cfg:      cfg,
	}, nil
}

func (tp *TUIProvider) GetProviderForFeature(
	ctx context.Context,
	feature capability.Feature,
	providerID string,
	market model.MarketType,
) (provider.Provider, model.MarketType, error) {
	opts := resolve.ResolutionOpts{
		Provider:   providerID,
		Market:     market,
		NoFallback: providerID != "",
	}

	p, actualMarket, wasFallback, err := tp.resolver.Resolve(ctx, feature, opts)
	if err != nil {
		return nil, "", err
	}

	_ = wasFallback

	return p, actualMarket, nil
}

func (tp *TUIProvider) GetAvailableProviders(feature capability.Feature) []string {
	var result []string

	for _, id := range registry.AllProviderIDs() {
		p, err := registry.NewProvider(id, tp.cfg)
		if err != nil {
			continue
		}

		caps := p.Capabilities()
		if caps[capability.CapabilityKey{Market: capability.MarketSpot, Feature: feature}] ||
			caps[capability.CapabilityKey{Market: capability.MarketFutures, Feature: feature}] {
			result = append(result, id)
		}
	}

	return result
}

func (tp *TUIProvider) GetAvailableMarkets(providerID string, feature capability.Feature) []model.MarketType {
	var result []model.MarketType

	p, err := registry.NewProvider(providerID, tp.cfg)
	if err != nil {
		return result
	}

	caps := p.Capabilities()

	for _, m := range []model.MarketType{model.MarketSpot, model.MarketFutures, model.MarketMargin} {
		if caps[capability.CapabilityKey{Market: m, Feature: feature}] {
			result = append(result, m)
		}
	}

	return result
}

func (tp *TUIProvider) Config() *config.Config {
	return tp.cfg
}

func (tp *TUIProvider) ID() string {
	return "tui-provider"
}
