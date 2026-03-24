package bitget

import (
	"testing"

	"github.com/mdnmdn/bits/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestBitgetClientIdentity(t *testing.T) {
	cfg := config.BitgetConfig{
		Key:    "test-key",
		Secret: "test-secret",
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
	tests := []struct {
		input    string
		expected string
	}{
		{"1m", "1min"},
		{"5m", "5min"},
		{"15m", "15min"},
		{"30m", "30min"},
		{"1h", "1h"},
		{"4h", "4h"},
		{"6h", "6h"},
		{"12h", "12h"},
		{"1d", "1day"},
		{"daily", "1day"},
		{"3d", "3day"},
		{"1w", "1week"},
		{"1M", "1M"},
		{"hourly", "1h"},
		{"invalid", "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, convertGranularityFormat(tt.input))
		})
	}
}
