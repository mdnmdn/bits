package resolve

import (
	"context"
	"fmt"

	"github.com/mdnmdn/bits/internal/capability"
	"github.com/mdnmdn/bits/internal/config"
	"github.com/mdnmdn/bits/internal/model"
	"github.com/mdnmdn/bits/internal/provider"
)

// ResolutionOpts controls how a provider is selected for a request.
type ResolutionOpts struct {
	Provider   string           // explicit provider id override ("" = use config/default)
	Market     model.MarketType // explicit market ("" = spot)
	NoFallback bool             // if true, error instead of fallback
}

// Resolver selects a provider for a given feature, applying fallback logic.
type Resolver struct {
	cfg      *config.Config
	registry func(name string, cfg *config.Config) (provider.Provider, error)
	all      func() []string
}

// New creates a Resolver backed by the given config and provider registry.
func New(cfg *config.Config,
	registry func(string, *config.Config) (provider.Provider, error),
	allIDs func() []string,
) *Resolver {
	return &Resolver{cfg: cfg, registry: registry, all: allIDs}
}

// Resolve picks the best provider+market for the requested feature.
// Returns: actualProvider, actualMarket, wasFallback, error.
func (r *Resolver) Resolve(
	ctx context.Context,
	feature capability.Feature,
	opts ResolutionOpts,
) (provider.Provider, model.MarketType, bool, error) {
	market := opts.Market
	if market == "" {
		market = model.MarketSpot
	}

	// Determine the requested provider id.
	requestedID := opts.Provider
	if requestedID == "" {
		requestedID = r.cfg.Provider
	}
	if requestedID == "" {
		requestedID = "coingecko"
	}

	// Try the requested provider first.
	p, err := r.registry(requestedID, r.cfg)
	if err != nil && opts.Provider != "" {
		return nil, "", false, err
	}
	if err == nil {
		key := capability.CapabilityKey{Market: market, Feature: feature}
		if p.Capabilities()[key] {
			return p, market, false, nil
		}
		// Requested market not supported — try spot as fallback market before
		// trying a different provider.
		if !opts.NoFallback && market != model.MarketSpot {
			if p.Capabilities()[capability.CapabilityKey{Market: model.MarketSpot, Feature: feature}] {
				return p, model.MarketSpot, true, nil
			}
		}
	}

	if opts.NoFallback {
		return nil, "", false, fmt.Errorf(
			"provider %q does not support feature %q for market %q", requestedID, feature, market,
		)
	}

	// Fallback: try all other registered providers.
	for _, id := range r.all() {
		if id == requestedID {
			continue
		}
		fp, ferr := r.registry(id, r.cfg)
		if ferr != nil {
			continue
		}
		key := capability.CapabilityKey{Market: market, Feature: feature}
		if fp.Capabilities()[key] {
			return fp, market, true, nil
		}
		// Also try spot.
		if market != model.MarketSpot {
			spotKey := capability.CapabilityKey{Market: model.MarketSpot, Feature: feature}
			if fp.Capabilities()[spotKey] {
				return fp, model.MarketSpot, true, nil
			}
		}
	}

	return nil, "", false, fmt.Errorf(
		"no provider supports feature %q for market %q", feature, market,
	)
}
