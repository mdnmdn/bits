// Package capability defines the provider capability matrix types.
// It is a standalone package (no internal project imports) so that
// provider sub-packages can import it without creating import cycles.
package capability

// MarketType represents a market segment supported by a provider.
type MarketType string

const (
	MarketSpot    MarketType = "spot"
	MarketFutures MarketType = "futures"
	MarketMargin  MarketType = "margin"
)

// Feature represents a data capability a provider may support.
type Feature string

const (
	FeaturePrice           Feature = "price"
	FeatureCandles         Feature = "candles"
	FeatureTicker24h       Feature = "ticker_24h"
	FeatureOrderBook       Feature = "order_book"
	FeatureCoinInfo        Feature = "coin_info"
	FeatureMarketsList     Feature = "markets_list"
	FeatureSearch          Feature = "search"
	FeatureTrending        Feature = "trending"
	FeatureHistorical      Feature = "historical"
	FeatureGainersLosers   Feature = "gainers_losers"
	FeatureStreamPrice     Feature = "stream_price"
	FeatureStreamOrderBook Feature = "stream_order_book"
)

// CapabilityKey uniquely identifies a (market, feature) combination.
type CapabilityKey struct {
	Market  MarketType
	Feature Feature
}

// CapabilityMatrix maps capability keys to whether the provider supports them.
type CapabilityMatrix map[CapabilityKey]bool

// CapabilityProvider is implemented by providers that declare their capability matrix.
type CapabilityProvider interface {
	Capabilities() CapabilityMatrix
}

// NewCapabilityMatrix builds a CapabilityMatrix with the given keys set to true.
func NewCapabilityMatrix(entries ...CapabilityKey) CapabilityMatrix {
	m := make(CapabilityMatrix, len(entries))
	for _, k := range entries {
		m[k] = true
	}
	return m
}

// AllFeatures returns features in a stable display order.
func AllFeatures() []Feature {
	return []Feature{
		FeaturePrice,
		FeatureCandles,
		FeatureTicker24h,
		FeatureOrderBook,
		FeatureCoinInfo,
		FeatureMarketsList,
		FeatureSearch,
		FeatureTrending,
		FeatureHistorical,
		FeatureGainersLosers,
		FeatureStreamPrice,
		FeatureStreamOrderBook,
	}
}

// AllMarkets returns market types in a stable display order.
func AllMarkets() []MarketType {
	return []MarketType{MarketSpot, MarketFutures, MarketMargin}
}
