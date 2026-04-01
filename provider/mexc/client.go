package mexc

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/mdnmdn/bits/internal/ws"
	"github.com/mdnmdn/bits/capability"
	"github.com/mdnmdn/bits/config"
	"github.com/mdnmdn/bits/model"
	"github.com/mdnmdn/bits/provider"
)

const (
	providerID     = "mexc"
	defaultBaseURL = "https://api.mexc.com"
	spotAPIPath    = "/api/v3"
	futuresAPIPath = "/api/v1/contract"
	spotWSURL      = "wss://wbs-api.mexc.com/ws"
	futuresWSURL   = "wss://contract.mexc.com/edge"
)

// Client is the MEXC provider implementation.
type Client struct {
	cfg        config.MEXCConfig
	httpClient *http.Client
	userAgent  string

	// Stream management
	streamMu    sync.RWMutex
	pricePool   *ws.Pool
	priceOut    <-chan ws.StreamResponse[any]
	priceChan   chan *model.CoinPrice
	priceSubs   map[string]bool
	priceStatus provider.StreamStatus
	priceMarket model.MarketType

	bookPool   *ws.Pool
	bookOut    <-chan ws.StreamResponse[any]
	bookChan   chan *model.OrderBook
	bookSubs   map[string]bool
	bookStatus provider.StreamStatus
	bookMarket model.MarketType
	bookDepth  int
}

// NewClient creates a new MEXC API client from the given config.
func NewClient(cfg config.MEXCConfig) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		priceChan:  make(chan *model.CoinPrice, 100),
		priceSubs:  make(map[string]bool),
		bookChan:   make(chan *model.OrderBook, 100),
		bookSubs:   make(map[string]bool),
	}
}

// ID returns the provider identifier.
func (c *Client) ID() string { return providerID }

// SetUserAgent sets the User-Agent header for API requests.
func (c *Client) SetUserAgent(ua string) { c.userAgent = ua }

// Capabilities returns the capability matrix for MEXC.
func (c *Client) Capabilities() capability.CapabilityMatrix {
	matrix := capability.CapabilityMatrix{}

	// Check if markets are enabled, default to spot if not configured
	spotEnabled := c.cfg.IsSpotEnabled()
	marginEnabled := c.cfg.IsMarginEnabled()
	futuresEnabled := c.cfg.IsFuturesEnabled()

	// Default to spot enabled if nothing explicitly enabled
	if !spotEnabled && !marginEnabled && !futuresEnabled {
		spotEnabled = true
		marginEnabled = true
		futuresEnabled = true
	}

	// Spot
	if spotEnabled {
		matrix[capability.CapabilityKey{Market: capability.MarketSpot, Feature: capability.FeatureServerTime}] = true
		matrix[capability.CapabilityKey{Market: capability.MarketSpot, Feature: capability.FeatureExchangeInfo}] = true
		matrix[capability.CapabilityKey{Market: capability.MarketSpot, Feature: capability.FeaturePrice}] = true
		matrix[capability.CapabilityKey{Market: capability.MarketSpot, Feature: capability.FeatureTicker24h}] = true
		matrix[capability.CapabilityKey{Market: capability.MarketSpot, Feature: capability.FeatureCandles}] = true
		matrix[capability.CapabilityKey{Market: capability.MarketSpot, Feature: capability.FeatureOrderBook}] = true
		matrix[capability.CapabilityKey{Market: capability.MarketSpot, Feature: capability.FeatureStreamPrice}] = true
		matrix[capability.CapabilityKey{Market: capability.MarketSpot, Feature: capability.FeatureStreamOrderBook}] = true
	}

	// Margin (served via Spot REST endpoints)
	if marginEnabled {
		matrix[capability.CapabilityKey{Market: capability.MarketMargin, Feature: capability.FeatureServerTime}] = true
		matrix[capability.CapabilityKey{Market: capability.MarketMargin, Feature: capability.FeatureExchangeInfo}] = true
		matrix[capability.CapabilityKey{Market: capability.MarketMargin, Feature: capability.FeaturePrice}] = true
		matrix[capability.CapabilityKey{Market: capability.MarketMargin, Feature: capability.FeatureTicker24h}] = true
		matrix[capability.CapabilityKey{Market: capability.MarketMargin, Feature: capability.FeatureCandles}] = true
		matrix[capability.CapabilityKey{Market: capability.MarketMargin, Feature: capability.FeatureOrderBook}] = true
	}

	// Futures
	if futuresEnabled {
		matrix[capability.CapabilityKey{Market: capability.MarketFutures, Feature: capability.FeatureServerTime}] = true
		matrix[capability.CapabilityKey{Market: capability.MarketFutures, Feature: capability.FeatureExchangeInfo}] = true
		matrix[capability.CapabilityKey{Market: capability.MarketFutures, Feature: capability.FeaturePrice}] = true
		matrix[capability.CapabilityKey{Market: capability.MarketFutures, Feature: capability.FeatureTicker24h}] = true
		matrix[capability.CapabilityKey{Market: capability.MarketFutures, Feature: capability.FeatureCandles}] = true
		matrix[capability.CapabilityKey{Market: capability.MarketFutures, Feature: capability.FeatureOrderBook}] = true
		matrix[capability.CapabilityKey{Market: capability.MarketFutures, Feature: capability.FeatureStreamPrice}] = true
		matrix[capability.CapabilityKey{Market: capability.MarketFutures, Feature: capability.FeatureStreamOrderBook}] = true
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
