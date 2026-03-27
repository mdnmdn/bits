package bitget

import "github.com/mdnmdn/bits/internal/capability"

func (c *Client) Capabilities() capability.CapabilityMatrix {
	s := capability.MarketSpot
	f := capability.MarketFutures

	enabledSpot := c.IsSpotEnabled()
	enabledFutures := c.IsFuturesEnabled()

	matrix := capability.CapabilityMatrix{}

	if enabledSpot {
		matrix[capability.CapabilityKey{Market: s, Feature: capability.FeaturePrice}] = true
		matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureCandles}] = true
		matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureTicker24h}] = true
	}

	if enabledFutures {
		matrix[capability.CapabilityKey{Market: f, Feature: capability.FeaturePrice}] = true
		matrix[capability.CapabilityKey{Market: f, Feature: capability.FeatureCandles}] = true
		matrix[capability.CapabilityKey{Market: f, Feature: capability.FeatureTicker24h}] = true
	}

	// If nothing enabled, default to spot capabilities
	if len(matrix) == 0 {
		matrix[capability.CapabilityKey{Market: s, Feature: capability.FeaturePrice}] = true
		matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureCandles}] = true
		matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureTicker24h}] = true
	}

	return matrix
}
