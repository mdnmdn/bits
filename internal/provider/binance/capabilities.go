package binance

import "github.com/mdnmdn/bits/internal/capability"

func (c *Client) Capabilities() capability.CapabilityMatrix {
	s := capability.MarketSpot
	m := capability.MarketMargin
	f := capability.MarketFutures

	enabledSpot := c.IsSpotEnabled()
	enabledMargin := c.IsMarginEnabled()
	enabledFutures := c.IsFuturesEnabled()

	matrix := capability.CapabilityMatrix{}

	if enabledSpot {
		matrix[capability.CapabilityKey{Market: s, Feature: capability.FeaturePrice}] = true
		matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureCandles}] = true
		matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureTicker24h}] = true
		matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureOrderBook}] = true
		matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureStreamOrderBook}] = true
	}

	if enabledMargin {
		matrix[capability.CapabilityKey{Market: m, Feature: capability.FeaturePrice}] = true
		matrix[capability.CapabilityKey{Market: m, Feature: capability.FeatureCandles}] = true
		matrix[capability.CapabilityKey{Market: m, Feature: capability.FeatureTicker24h}] = true
	}

	if enabledFutures {
		matrix[capability.CapabilityKey{Market: f, Feature: capability.FeaturePrice}] = true
		matrix[capability.CapabilityKey{Market: f, Feature: capability.FeatureCandles}] = true
		matrix[capability.CapabilityKey{Market: f, Feature: capability.FeatureTicker24h}] = true
		matrix[capability.CapabilityKey{Market: f, Feature: capability.FeatureOrderBook}] = true
	}

	// If nothing enabled, default to spot capabilities
	if len(matrix) == 0 {
		matrix[capability.CapabilityKey{Market: s, Feature: capability.FeaturePrice}] = true
		matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureCandles}] = true
		matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureTicker24h}] = true
		matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureOrderBook}] = true
		matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureStreamOrderBook}] = true
	}

	return matrix
}
