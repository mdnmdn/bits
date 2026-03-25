package binance

import (
	binance "github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/mdnmdn/bits/internal/config"
)

const providerID = "binance"

// Client represents a Binance API client that implements the Provider,
// TickerProvider, and OrderBookProvider interfaces.
type Client struct {
	config        config.BinanceConfig
	marketType    string
	client        *binance.Client
	futuresClient *futures.Client
	userAgent     string
}

// NewClient creates a new Binance provider client.
func NewClient(cfg config.BinanceConfig) *Client {
	var client *binance.Client
	var futuresClient *futures.Client

	// Default to spot if not specified
	marketType := cfg.MarketType
	if marketType == "" {
		marketType = config.MarketTypeSpot
	}

	if marketType == config.MarketTypeFuture {
		futuresClient = futures.NewClient(cfg.APIKey, cfg.APISecret)
		if cfg.UseTestnet {
			futuresClient.BaseURL = "https://testnet.binancefuture.com"
		}
	} else {
		// Spot and Margin use the standard client
		client = binance.NewClient(cfg.APIKey, cfg.APISecret)
		if cfg.UseTestnet {
			client.BaseURL = "https://testnet.binance.vision"
		} else if cfg.BaseURL != "" {
			client.BaseURL = cfg.BaseURL
		}
	}

	return &Client{
		config:        cfg,
		marketType:    marketType,
		client:        client,
		futuresClient: futuresClient,
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
