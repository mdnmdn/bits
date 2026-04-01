package mexc

import (
	"testing"

	"github.com/mdnmdn/bits/pkg/config"
	"github.com/mdnmdn/bits/pkg/provider"
	"github.com/stretchr/testify/assert"
)

func TestMEXCClient_ID(t *testing.T) {
	cfg := config.MEXCConfig{}
	client := NewClient(cfg)
	assert.Equal(t, "mexc", client.ID())
}

func TestMEXCClient_Capabilities(t *testing.T) {
	cfg := config.MEXCConfig{}
	client := NewClient(cfg)
	caps := client.Capabilities()
	assert.NotNil(t, caps)
	assert.True(t, len(caps) > 0)
}

var _ provider.Provider = (*Client)(nil)
var _ provider.ExchangeProvider = (*Client)(nil)
var _ provider.PriceProvider = (*Client)(nil)
var _ provider.CandleProvider = (*Client)(nil)
var _ provider.TickerProvider = (*Client)(nil)
var _ provider.OrderBookProvider = (*Client)(nil)
var _ provider.PriceStreamProvider = (*Client)(nil)
var _ provider.OrderBookStreamProvider = (*Client)(nil)
