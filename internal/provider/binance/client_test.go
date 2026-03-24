package binance

import (
	"testing"

	"github.com/coingecko/coingecko-cli/internal/config"
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

	t.Run("Testnet BaseURL", func(t *testing.T) {
		cfg := config.BinanceConfig{UseTestnet: true}
		client := NewClient(cfg)
		assert.Equal(t, "https://testnet.binance.vision", client.GetClient().BaseURL)
	})

	t.Run("Custom BaseURL", func(t *testing.T) {
		customURL := "https://custom.binance.com"
		cfg := config.BinanceConfig{BaseURL: customURL}
		client := NewClient(cfg)
		assert.Equal(t, customURL, client.GetClient().BaseURL)
	})
}
