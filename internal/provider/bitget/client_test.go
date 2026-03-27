package bitget

import (
	"testing"

	"github.com/mdnmdn/bits/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestBitgetClientIdentity(t *testing.T) {
	cfg := config.BitgetConfig{
		APIKey:    "test-key",
		APISecret: "test-secret",
	}
	client := NewClient(cfg)

	assert.Equal(t, "bitget", client.ID())

	ua := "test-user-agent"
	client.SetUserAgent(ua)
	assert.Equal(t, ua, client.userAgent)
}

func TestBitgetClientConfiguration(t *testing.T) {
	t.Run("Default BaseURL", func(t *testing.T) {
		cfg := config.BitgetConfig{}
		client := NewClient(cfg)
		assert.Equal(t, DefaultBaseURL, client.config.BaseURL)
	})

	t.Run("Custom BaseURL", func(t *testing.T) {
		customURL := "https://custom.bitget.com"
		cfg := config.BitgetConfig{BaseURL: customURL}
		client := NewClient(cfg)
		assert.Equal(t, customURL, client.config.BaseURL)
	})
}

func TestConvertGranularityFormat(t *testing.T) {
	spotTests := []struct {
		input    string
		expected string
	}{
		{"1m", "1min"},
		{"5m", "5min"},
		{"1h", "1h"},
		{"4h", "4h"},
		{"1d", "1day"},
		{"daily", "1day"},
		{"hourly", "1h"},
		{"invalid", "invalid"},
	}

	for _, tt := range spotTests {
		t.Run("spot_"+tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, convertGranularityFormat(tt.input, "spot"))
		})
	}

	futureTests := []struct {
		input    string
		expected string
	}{
		{"1m", "1m"},
		{"5m", "5m"},
		{"1h", "1H"},
		{"4h", "4H"},
		{"1d", "1D"},
		{"daily", "1D"},
		{"hourly", "1H"},
		{"invalid", "invalid"},
	}

	for _, tt := range futureTests {
		t.Run("future_"+tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, convertGranularityFormat(tt.input, "future"))
		})
	}
}
