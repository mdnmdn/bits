package binance

import (
	binance "github.com/adshao/go-binance/v2"
	"github.com/coingecko/coingecko-cli/internal/config"
)

const providerID = "binance"

// Client represents a Binance API client that implements the Provider,
// TickerProvider, and OrderBookProvider interfaces.
type Client struct {
	config    config.BinanceConfig
	client    *binance.Client
	userAgent string
}

// NewClient creates a new Binance provider client.
func NewClient(cfg config.BinanceConfig) *Client {
	client := binance.NewClient(cfg.APIKey, cfg.APISecret)

	if cfg.UseTestnet {
		client.BaseURL = "https://testnet.binance.vision"
	} else if cfg.BaseURL != "" {
		client.BaseURL = cfg.BaseURL
	}

	return &Client{
		config: cfg,
		client: client,
	}
}

// ID returns the provider identifier.
func (c *Client) ID() string {
	return providerID
}

// SetUserAgent sets the User-Agent string for HTTP requests.
func (c *Client) SetUserAgent(userAgent string) {
	c.userAgent = userAgent
}

// GetClient returns the underlying go-binance client for advanced usage.
func (c *Client) GetClient() *binance.Client {
	return c.client
}
