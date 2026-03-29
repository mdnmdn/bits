package bitget

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/mdnmdn/bits/internal/auth"
	"github.com/mdnmdn/bits/pkg/capability"
	"github.com/mdnmdn/bits/pkg/config"
)

const (
	providerID     = "bitget"
	defaultBaseURL = "https://api.bitget.com"
)

// Client is the new Bitget provider implementation.
// Market type is passed per-method, not stored in client state.
type Client struct {
	cfg        config.BitgetConfig
	httpClient *http.Client
	userAgent  string
}

// NewClient creates a new Bitget API client from the given config.
func NewClient(cfg config.BitgetConfig) *Client {
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

// Capabilities returns the capability matrix based on the configured markets.
// Spot is enabled if cfg.IsSpotEnabled() or nothing else is enabled (default spot).
func (c *Client) Capabilities() capability.CapabilityMatrix {
	s := capability.MarketSpot
	f := capability.MarketFutures
	m := capability.MarketMargin

	spotEnabled := c.cfg.IsSpotEnabled()
	futuresEnabled := c.cfg.IsFuturesEnabled()
	marginEnabled := c.cfg.IsMarginEnabled()

	// Default to spot if nothing explicitly enabled
	if !spotEnabled && !futuresEnabled && !marginEnabled {
		spotEnabled = true
	}

	matrix := capability.CapabilityMatrix{}

	// ServerTime is provider-level; register under spot.
	matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureServerTime}] = true

	if spotEnabled {
		matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureExchangeInfo}] = true
		matrix[capability.CapabilityKey{Market: s, Feature: capability.FeaturePrice}] = true
		matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureCandles}] = true
		matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureTicker24h}] = true
		matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureOrderBook}] = true
		matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureStreamPrice}] = true
		matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureStreamOrderBook}] = true
	}

	if futuresEnabled {
		matrix[capability.CapabilityKey{Market: f, Feature: capability.FeatureExchangeInfo}] = true
		matrix[capability.CapabilityKey{Market: f, Feature: capability.FeatureCandles}] = true
		matrix[capability.CapabilityKey{Market: f, Feature: capability.FeatureTicker24h}] = true
		matrix[capability.CapabilityKey{Market: f, Feature: capability.FeatureOrderBook}] = true
		matrix[capability.CapabilityKey{Market: f, Feature: capability.FeatureStreamPrice}] = true
		matrix[capability.CapabilityKey{Market: f, Feature: capability.FeatureStreamOrderBook}] = true
	}

	if marginEnabled {
		matrix[capability.CapabilityKey{Market: m, Feature: capability.FeatureExchangeInfo}] = true
		matrix[capability.CapabilityKey{Market: m, Feature: capability.FeatureTicker24h}] = true
	}

	return matrix
}

// doRequest performs an authenticated GET request to the Bitget API.
// path is the endpoint path (e.g. "/api/v2/public/time"), query is the raw query string (without "?").
func (c *Client) doRequest(method, path, query string) ([]byte, error) {
	fullPath := path
	if query != "" {
		fullPath = path + "?" + query
	}

	url := c.cfg.BaseURL + fullPath

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	// Signature message: timestamp + METHOD + fullPath + body (empty for GET)
	message := timestamp + method + fullPath
	sign := auth.GenerateHMACSHA256Base64(message, c.cfg.APISecret)

	req.Header.Set("ACCESS-KEY", c.cfg.APIKey)
	req.Header.Set("ACCESS-SIGN", sign)
	req.Header.Set("ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("ACCESS-PASSPHRASE", c.cfg.Passphrase)
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
