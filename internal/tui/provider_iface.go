package tui

import (
	"context"

	"github.com/mdnmdn/bits/pkg/capability"
	"github.com/mdnmdn/bits/pkg/model"
	"github.com/mdnmdn/bits/pkg/provider"
)

type ProviderWrapper interface {
	GetProviderForFeature(ctx context.Context, feature capability.Feature, providerID string, market model.MarketType) (provider.Provider, model.MarketType, error)
	GetAvailableProviders(feature capability.Feature) []string
	GetAvailableMarkets(providerID string, feature capability.Feature) []model.MarketType
}
