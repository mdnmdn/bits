package cryptocom

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mdnmdn/bits/pkg/capability"
	"github.com/mdnmdn/bits/pkg/config"
)

const (
	providerID     = "cryptocom"
	defaultBaseURL = "https://api.crypto.com/v2"
)

// Client is the Crypto.com Exchange provider implementation.
// Only spot market is supported via the v2 REST API.
type Client struct {
	cfg        config.CryptoComConfig
	httpClient *http.Client
	userAgent  string
}

// NewClient creates a new Crypto.com API client from the given config.
func NewClient(cfg config.CryptoComConfig) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// ID returns the provider identifier.
func (c *Client) ID() string { return providerID }

// SetUserAgent sets the User-Agent header for API requests.
func (c *Client) SetUserAgent(ua string) { c.userAgent = ua }

// Capabilities returns the capability matrix for Crypto.com.
// REST and WebSocket streaming are supported for the spot market.
func (c *Client) Capabilities() capability.CapabilityMatrix {
	s := capability.MarketSpot
	return capability.NewCapabilityMatrix(
		capability.CapabilityKey{Market: s, Feature: capability.FeatureServerTime},
		capability.CapabilityKey{Market: s, Feature: capability.FeatureExchangeInfo},
		capability.CapabilityKey{Market: s, Feature: capability.FeaturePrice},
		capability.CapabilityKey{Market: s, Feature: capability.FeatureTicker24h},
		capability.CapabilityKey{Market: s, Feature: capability.FeatureOrderBook},
		capability.CapabilityKey{Market: s, Feature: capability.FeatureStreamPrice},
		capability.CapabilityKey{Market: s, Feature: capability.FeatureStreamOrderBook},
	)
}

// doRequest performs a GET request to the Crypto.com v2 API.
// path is the endpoint path without a leading slash (e.g. "public/get-ticker").
// query is an optional raw query string without the leading "?".
func (c *Client) doRequest(path, query string) ([]byte, error) {
	fullPath := path
	if query != "" {
		fullPath = path + "?" + query
	}
	url := c.cfg.BaseURL + "/" + fullPath

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return body, nil
}
