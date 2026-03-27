package provider

import "github.com/mdnmdn/bits/internal/capability"

// Re-export capability types so callers can use provider.* names.

type MarketType = capability.MarketType
type Feature = capability.Feature
type CapabilityKey = capability.CapabilityKey
type CapabilityMatrix = capability.CapabilityMatrix
type CapabilityProvider = capability.CapabilityProvider

const (
	MarketSpot    = capability.MarketSpot
	MarketFutures = capability.MarketFutures
	MarketMargin  = capability.MarketMargin

	FeaturePrice           = capability.FeaturePrice
	FeatureCandles         = capability.FeatureCandles
	FeatureTicker24h       = capability.FeatureTicker24h
	FeatureOrderBook       = capability.FeatureOrderBook
	FeatureCoinInfo        = capability.FeatureCoinInfo
	FeatureMarketsList     = capability.FeatureMarketsList
	FeatureSearch          = capability.FeatureSearch
	FeatureTrending        = capability.FeatureTrending
	FeatureHistorical      = capability.FeatureHistorical
	FeatureGainersLosers   = capability.FeatureGainersLosers
	FeatureStreamPrice     = capability.FeatureStreamPrice
	FeatureStreamOrderBook = capability.FeatureStreamOrderBook
)

func NewCapabilityMatrix(entries ...CapabilityKey) CapabilityMatrix {
	return capability.NewCapabilityMatrix(entries...)
}

func AllFeatures() []Feature    { return capability.AllFeatures() }
func AllMarkets() []MarketType  { return capability.AllMarkets() }
