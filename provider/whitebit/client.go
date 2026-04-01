package whitebit

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/mdnmdn/bits/capability"
	"github.com/mdnmdn/bits/config"
	"github.com/mdnmdn/bits/internal/ws"
	"github.com/mdnmdn/bits/model"
	"github.com/mdnmdn/bits/provider"
)

const (
	providerID     = "whitebit"
	defaultBaseURL = "https://whitebit.com"
	wsURLRetry     = "wss://api.whitebit.eu/ws"
)

var _ provider.Provider = (*Client)(nil)
var _ provider.PriceStreamProvider = (*Client)(nil)
var _ provider.OrderBookStreamProvider = (*Client)(nil)

// Client is the WhiteBit provider implementation.
type Client struct {
	cfg        config.WhiteBitConfig
	httpClient *http.Client
	userAgent  string

	pricePool   *ws.Pool
	priceOut    <-chan ws.StreamResponse[any]
	priceChan   chan *model.CoinPrice
	priceSubs   map[string]bool
	priceStatus provider.StreamStatus
	streamMu    sync.RWMutex

	bookPool   *ws.Pool
	bookOut    <-chan ws.StreamResponse[any]
	bookChan   chan *model.OrderBook
	bookSubs   map[string]bool
	bookStatus provider.StreamStatus
	market     model.MarketType

	// priceMerged stores merged state for dual-subscription price stream
	priceMerged map[string]*mergedPriceState
}

// NewClient creates a new WhiteBit API client from the given config.
func NewClient(cfg config.WhiteBitConfig) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	return &Client{
		cfg:         cfg,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		priceChan:   make(chan *model.CoinPrice, 100),
		priceSubs:   make(map[string]bool),
		priceStatus: provider.StreamStatusStopped,
		bookChan:    make(chan *model.OrderBook, 100),
		bookSubs:    make(map[string]bool),
		bookStatus:  provider.StreamStatusStopped,
	}
}

// ID returns the provider identifier.
func (c *Client) ID() string { return providerID }

// SetUserAgent sets the User-Agent header for API requests.
func (c *Client) SetUserAgent(ua string) { c.userAgent = ua }

// Capabilities returns the capability matrix for WhiteBit.
func (c *Client) Capabilities() capability.CapabilityMatrix {
	s := capability.MarketSpot
	f := capability.MarketFutures

	matrix := capability.CapabilityMatrix{}

	// Register all spot features
	matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureServerTime}] = true
	matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureExchangeInfo}] = true
	matrix[capability.CapabilityKey{Market: s, Feature: capability.FeaturePrice}] = true
	matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureCandles}] = true
	matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureTicker24h}] = true
	matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureOrderBook}] = true

	// Register futures features (Time and Info are shared)
	matrix[capability.CapabilityKey{Market: f, Feature: capability.FeatureExchangeInfo}] = true
	matrix[capability.CapabilityKey{Market: f, Feature: capability.FeatureCandles}] = true
	matrix[capability.CapabilityKey{Market: f, Feature: capability.FeatureTicker24h}] = true
	matrix[capability.CapabilityKey{Market: f, Feature: capability.FeatureOrderBook}] = true

	// Register streaming features (Spot and Futures)
	matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureStreamPrice}] = true
	matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureStreamOrderBook}] = true
	matrix[capability.CapabilityKey{Market: f, Feature: capability.FeatureStreamPrice}] = true
	matrix[capability.CapabilityKey{Market: f, Feature: capability.FeatureStreamOrderBook}] = true

	return matrix
}

// doRequest performs a plain GET request to the WhiteBit API.
func (c *Client) doRequest(path string) ([]byte, error) {
	url := c.cfg.BaseURL + path

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
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return body, nil
}
