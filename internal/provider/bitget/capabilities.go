package bitget

import "github.com/mdnmdn/bits/internal/capability"

func (c *Client) Capabilities() capability.CapabilityMatrix {
	s := capability.MarketSpot
	f := capability.MarketFutures
	return capability.NewCapabilityMatrix(
		// Spot
		capability.CapabilityKey{Market: s, Feature: capability.FeaturePrice},
		capability.CapabilityKey{Market: s, Feature: capability.FeatureCandles},
		capability.CapabilityKey{Market: s, Feature: capability.FeatureTicker24h},
		// Futures
		capability.CapabilityKey{Market: f, Feature: capability.FeaturePrice},
		capability.CapabilityKey{Market: f, Feature: capability.FeatureCandles},
		capability.CapabilityKey{Market: f, Feature: capability.FeatureTicker24h},
	)
}
