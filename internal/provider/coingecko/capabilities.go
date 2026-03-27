package coingecko

import "github.com/mdnmdn/bits/internal/capability"

func (c *Client) Capabilities() capability.CapabilityMatrix {
	s := capability.MarketSpot
	return capability.NewCapabilityMatrix(
		capability.CapabilityKey{Market: s, Feature: capability.FeaturePrice},
		capability.CapabilityKey{Market: s, Feature: capability.FeatureCandles},
		capability.CapabilityKey{Market: s, Feature: capability.FeatureCoinInfo},
		capability.CapabilityKey{Market: s, Feature: capability.FeatureMarketsList},
		capability.CapabilityKey{Market: s, Feature: capability.FeatureSearch},
		capability.CapabilityKey{Market: s, Feature: capability.FeatureTrending},
		capability.CapabilityKey{Market: s, Feature: capability.FeatureHistorical},
		capability.CapabilityKey{Market: s, Feature: capability.FeatureGainersLosers},
		capability.CapabilityKey{Market: s, Feature: capability.FeatureStreamPrice},
	)
}
