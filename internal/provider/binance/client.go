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
	config           config.BinanceConfig
	activeMarketType string
	spotEnabled      bool
	marginEnabled    bool
	futuresEnabled   bool
	client           *binance.Client
	futuresClient    *futures.Client
	userAgent        string
}

// NewClient creates a new Binance provider client.
func NewClient(cfg config.BinanceConfig) *Client {
	spotEnabled := cfg.IsSpotEnabled()
	marginEnabled := cfg.IsMarginEnabled()
	futuresEnabled := cfg.IsFuturesEnabled()

	// Default to spot if nothing explicitly enabled
	if !spotEnabled && !marginEnabled && !futuresEnabled {
		spotEnabled = true // default
	}

	var client *binance.Client
	var futuresClient *futures.Client

	// Initialize spot client if spot or margin enabled
	if spotEnabled || marginEnabled {
		client = binance.NewClient(cfg.APIKey, cfg.APISecret)
		if cfg.BaseURL != "" {
			client.BaseURL = cfg.BaseURL
		}
	}

	// Initialize futures client if futures enabled
	if futuresEnabled {
		futuresClient = futures.NewClient(cfg.APIKey, cfg.APISecret)
		if cfg.Futures.UseTestnet {
			futuresClient.BaseURL = "https://testnet.binancefuture.com"
		}
	}

	// Determine active market (priority: futures > margin > spot)
	activeMarket := config.MarketTypeSpot
	if futuresEnabled {
		activeMarket = config.MarketTypeFuture
	} else if marginEnabled {
		activeMarket = config.MarketTypeMargin
	}

	return &Client{
		config:           cfg,
		activeMarketType: activeMarket,
		spotEnabled:      spotEnabled,
		marginEnabled:    marginEnabled,
		futuresEnabled:   futuresEnabled,
		client:           client,
		futuresClient:    futuresClient,
	}
}

// SetMarketType sets the active market type for subsequent API calls.
func (c *Client) SetMarketType(marketType string) {
	if marketType == config.MarketTypeFuture && c.futuresClient != nil {
		c.activeMarketType = config.MarketTypeFuture
	} else if marketType == config.MarketTypeMargin && c.client != nil {
		c.activeMarketType = config.MarketTypeMargin
	} else if c.client != nil {
		c.activeMarketType = config.MarketTypeSpot
	}
}

// MarketType returns the currently active market type.
func (c *Client) MarketType() string {
	return c.activeMarketType
}

// IsSpotEnabled returns true if spot market is configured.
func (c *Client) IsSpotEnabled() bool {
	return c.spotEnabled
}

// IsMarginEnabled returns true if margin market is configured.
func (c *Client) IsMarginEnabled() bool {
	return c.marginEnabled
}

// IsFuturesEnabled returns true if futures market is configured.
func (c *Client) IsFuturesEnabled() bool {
	return c.futuresEnabled
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
