package coingecko

import (
	"context"
	"encoding/json"
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

const providerID = "coingecko"

// Client is a thin CoinGecko provider implementing the new provider interfaces.
type Client struct {
	cfg       *config.Config
	http      *http.Client
	UserAgent string

	// Stream management
	streamMu      sync.RWMutex
	wsClient      *ws.Client
	priceChan     chan *model.CoinPrice
	priceSubsList []string
	priceStatus   provider.StreamStatus
}

// NewClient creates a new CoinGecko provider client.
func NewClient(cfg *config.Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: 30 * time.Second},
	}
}

// ID returns the provider identifier.
func (c *Client) ID() string { return providerID }

// SetUserAgent sets the User-Agent header for all requests.
func (c *Client) SetUserAgent(ua string) { c.UserAgent = ua }

// Capabilities returns the capability matrix for this provider.
// CoinGecko supports all spot features.
func (c *Client) Capabilities() capability.CapabilityMatrix {
	return capability.NewCapabilityMatrix(
		capability.CapabilityKey{Market: capability.MarketSpot, Feature: capability.FeaturePrice},
		capability.CapabilityKey{Market: capability.MarketSpot, Feature: capability.FeatureCandles},
		capability.CapabilityKey{Market: capability.MarketSpot, Feature: capability.FeatureMarketsList},
		capability.CapabilityKey{Market: capability.MarketSpot, Feature: capability.FeatureStreamPrice},
	)
}

// get makes an authenticated GET request and decodes the JSON response.
func (c *Client) get(ctx context.Context, path string, result any) error {
	if c.cfg.CoinGecko.APIKey == "" {
		return providerErr(model.ErrKindInvalidRequest,
			"CoinGecko API key not configured — set coingecko.api_key in config or BITS_COINGECKO_API_KEY env var (free demo key at coingecko.com/api)",
			nil)
	}

	url := c.cfg.CoinGecko.GetBaseURL() + path

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return providerErr(model.ErrKindUnknown, "creating request", err)
	}

	key, val := c.cfg.CoinGecko.GetAuthHeader()
	req.Header.Set(key, val)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		// Check for context errors first
		if cerr := contextErr(err); cerr != nil {
			return cerr
		}
		// Check for network errors
		if isNetworkErr(err) {
			return providerErr(model.ErrKindNetwork, "network error", err)
		}
		// Generic unknown error
		return providerErr(model.ErrKindUnknown, "HTTP request failed", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return httpErr(resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return providerErr(model.ErrKindParse, "decoding JSON response", err)
	}
	return nil
}
