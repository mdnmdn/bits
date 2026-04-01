package binance

import (
	goBinance "github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/mdnmdn/bits/capability"
	"github.com/mdnmdn/bits/config"
)

const providerID = "binance"

// Client implements the Binance provider for the new provider interfaces.
// Market type is passed as a method parameter; it is NOT stored in client state.
type Client struct {
	cfg           config.BinanceConfig
	spotClient    *goBinance.Client // nil if spot/margin not configured
	futuresClient *futures.Client   // nil if futures not configured
	userAgent     string
}

// NewClient creates a new Binance provider client.
func NewClient(cfg config.BinanceConfig) *Client {
	spotEnabled := cfg.IsSpotEnabled()
	marginEnabled := cfg.IsMarginEnabled()
	futuresEnabled := cfg.IsFuturesEnabled()

	// Default to spot if nothing explicitly enabled
	if !spotEnabled && !marginEnabled && !futuresEnabled {
		spotEnabled = true
	}

	var spotClient *goBinance.Client
	var futuresClient *futures.Client

	if spotEnabled || marginEnabled {
		spotClient = goBinance.NewClient(cfg.APIKey, cfg.APISecret)
		if cfg.BaseURL != "" {
			spotClient.BaseURL = cfg.BaseURL
		}
	}

	if futuresEnabled {
		futuresClient = futures.NewClient(cfg.APIKey, cfg.APISecret)
		if cfg.Futures.UseTestnet {
			futuresClient.BaseURL = "https://testnet.binancefuture.com"
		}
	}

	return &Client{
		cfg:           cfg,
		spotClient:    spotClient,
		futuresClient: futuresClient,
	}
}

// ID returns the provider identifier.
func (c *Client) ID() string { return providerID }

// SetUserAgent sets the User-Agent string for HTTP requests.
func (c *Client) SetUserAgent(ua string) { c.userAgent = ua }

// Capabilities returns the capability matrix for this client.
// Spot features are always enabled. Futures features enabled when futures is configured.
// Margin features (price, candles, ticker) enabled when margin is configured.
func (c *Client) Capabilities() capability.CapabilityMatrix {
	s := capability.MarketSpot
	m := capability.MarketMargin
	f := capability.MarketFutures

	spotEnabled := c.cfg.IsSpotEnabled() || c.spotClient != nil
	marginEnabled := c.cfg.IsMarginEnabled()
	futuresEnabled := c.cfg.IsFuturesEnabled()

	// If nothing is enabled, default to spot
	if !spotEnabled && !marginEnabled && !futuresEnabled {
		spotEnabled = true
	}

	keys := make([]capability.CapabilityKey, 0, 20)

	// ServerTime and ExchangeInfo are market-agnostic exchange features; register under spot.
	keys = append(keys,
		capability.CapabilityKey{Market: s, Feature: capability.FeatureServerTime},
		capability.CapabilityKey{Market: s, Feature: capability.FeatureExchangeInfo},
		capability.CapabilityKey{Market: f, Feature: capability.FeatureExchangeInfo},
	)

	if spotEnabled {
		keys = append(keys,
			capability.CapabilityKey{Market: s, Feature: capability.FeaturePrice},
			capability.CapabilityKey{Market: s, Feature: capability.FeatureCandles},
			capability.CapabilityKey{Market: s, Feature: capability.FeatureTicker24h},
			capability.CapabilityKey{Market: s, Feature: capability.FeatureOrderBook},
			capability.CapabilityKey{Market: s, Feature: capability.FeatureStreamPrice},
			capability.CapabilityKey{Market: s, Feature: capability.FeatureStreamOrderBook},
		)
	}

	if futuresEnabled {
		keys = append(keys,
			capability.CapabilityKey{Market: f, Feature: capability.FeaturePrice},
			capability.CapabilityKey{Market: f, Feature: capability.FeatureCandles},
			capability.CapabilityKey{Market: f, Feature: capability.FeatureTicker24h},
			capability.CapabilityKey{Market: f, Feature: capability.FeatureOrderBook},
			capability.CapabilityKey{Market: f, Feature: capability.FeatureStreamOrderBook},
		)
	}

	if marginEnabled {
		keys = append(keys,
			capability.CapabilityKey{Market: m, Feature: capability.FeatureExchangeInfo},
			capability.CapabilityKey{Market: m, Feature: capability.FeaturePrice},
			capability.CapabilityKey{Market: m, Feature: capability.FeatureCandles},
			capability.CapabilityKey{Market: m, Feature: capability.FeatureTicker24h},
		)
	}

	return capability.NewCapabilityMatrix(keys...)
}
