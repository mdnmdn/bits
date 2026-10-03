package okx

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mdnmdn/bits/capability"
	"github.com/mdnmdn/bits/config"
	"github.com/mdnmdn/bits/provider"
)

var (
	_ provider.Provider         = (*Client)(nil)
	_ provider.ExchangeProvider = (*Client)(nil)
	_ provider.CandleProvider   = (*Client)(nil)
	_ provider.TickerProvider   = (*Client)(nil)
)

func TestDefaultBaseURLIsEEA(t *testing.T) {
	assert.Equal(t, "https://eea.okx.com", NewClient(config.OKXConfig{}).cfg.BaseURL)
	assert.Equal(t, "https://www.okx.com", NewClient(config.OKXConfig{BaseURL: "https://www.okx.com"}).cfg.BaseURL)
}

func TestCapabilities(t *testing.T) {
	m := NewClient(config.OKXConfig{}).Capabilities()
	assert.True(t, m[capability.CapabilityKey{Market: capability.MarketSpot, Feature: capability.FeatureCandles}])
	assert.False(t, m[capability.CapabilityKey{Market: capability.MarketFutures, Feature: capability.FeatureCandles}])

	cfg := config.OKXConfig{Futures: config.MarketConfig{Enabled: true}}
	m = NewClient(cfg).Capabilities()
	assert.True(t, m[capability.CapabilityKey{Market: capability.MarketFutures, Feature: capability.FeatureExchangeInfo}])
	assert.True(t, m[capability.CapabilityKey{Market: capability.MarketFutures, Feature: capability.FeatureTicker24h}])
}
