package binance

import (
	"testing"

	"github.com/mdnmdn/bits/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestBinanceClientIdentity(t *testing.T) {
	cfg := config.BinanceConfig{
		APIKey:    "test-key",
		APISecret: "test-secret",
	}
	client := NewClient(cfg)

	assert.Equal(t, "binance", client.ID())

	ua := "test-user-agent"
	client.SetUserAgent(ua)
	assert.Equal(t, ua, client.userAgent)
}

func TestBinanceClientConfiguration(t *testing.T) {
	t.Run("Default BaseURL", func(t *testing.T) {
		cfg := config.BinanceConfig{}
		client := NewClient(cfg)
		assert.Equal(t, "https://api.binance.com", client.GetClient().BaseURL)
	})

	t.Run("Spot Enabled by Default", func(t *testing.T) {
		cfg := config.BinanceConfig{}
		client := NewClient(cfg)
		assert.True(t, client.IsSpotEnabled())
		assert.False(t, client.IsFuturesEnabled())
	})

	t.Run("Futures with Testnet", func(t *testing.T) {
		cfg := config.BinanceConfig{
			Futures: config.MarketConfig{
				Enabled:    true,
				UseTestnet: true,
			},
		}
		client := NewClient(cfg)
		assert.True(t, client.IsFuturesEnabled())
	})

	t.Run("Custom BaseURL", func(t *testing.T) {
		customURL := "https://custom.binance.com"
		cfg := config.BinanceConfig{BaseURL: customURL}
		client := NewClient(cfg)
		assert.Equal(t, customURL, client.GetClient().BaseURL)
	})
}
