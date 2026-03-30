package mexc

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mdnmdn/bits/pkg/capability"
	"github.com/mdnmdn/bits/pkg/config"
	"github.com/mdnmdn/bits/pkg/model"
)

const (
	providerID     = "mexc"
	defaultBaseURL = "https://api.mexc.com"
	spotAPIPath    = "/api/v3"
	futuresAPIPath = "/api/v1/contract"
)

// Client is the MEXC provider implementation.
type Client struct {
	cfg        config.MEXCConfig
	httpClient *http.Client
	userAgent  string
}

// NewClient creates a new MEXC API client from the given config.
func NewClient(cfg config.MEXCConfig) *Client {
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

// Capabilities returns the capability matrix for MEXC.
func (c *Client) Capabilities() capability.CapabilityMatrix {
	matrix := capability.CapabilityMatrix{}

	markets := []capability.MarketType{
		capability.MarketSpot,
		capability.MarketMargin,
		capability.MarketFutures,
	}

	for _, m := range markets {
		// Note: MEXC Margin market data is served via Spot REST endpoints.
		// All markets support these features via REST
		matrix[capability.CapabilityKey{Market: m, Feature: capability.FeatureServerTime}] = true
		matrix[capability.CapabilityKey{Market: m, Feature: capability.FeatureExchangeInfo}] = true
		matrix[capability.CapabilityKey{Market: m, Feature: capability.FeaturePrice}] = true
		matrix[capability.CapabilityKey{Market: m, Feature: capability.FeatureTicker24h}] = true
		matrix[capability.CapabilityKey{Market: m, Feature: capability.FeatureCandles}] = true
		matrix[capability.CapabilityKey{Market: m, Feature: capability.FeatureOrderBook}] = true
	}

	return matrix
}

// doRequest performs a GET request to the MEXC API.
func (c *Client) doRequest(ctx context.Context, market model.MarketType, path string, query string) ([]byte, error) {
	var fullURL string
	if market == model.MarketFutures {
		fullURL = c.cfg.BaseURL + futuresAPIPath + path
	} else {
		// Spot and Margin use the same base API path
		fullURL = c.cfg.BaseURL + spotAPIPath + path
	}

	if query != "" {
		fullURL += "?" + query
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	// Add API Key if available
	if c.cfg.APIKey != "" {
		req.Header.Set("X-MEXC-APIKEY", c.cfg.APIKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
