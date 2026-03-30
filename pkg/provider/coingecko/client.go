package coingecko

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mdnmdn/bits/pkg/capability"
	"github.com/mdnmdn/bits/pkg/config"
)

const providerID = "coingecko"

// Client is a thin CoinGecko provider implementing the new provider interfaces.
type Client struct {
	cfg       *config.Config
	http      *http.Client
	UserAgent string
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
		return fmt.Errorf("CoinGecko API key not configured — set coingecko.api_key in config or BITS_COINGECKO_API_KEY env var (free demo key at coingecko.com/api)")
	}

	url := c.cfg.CoinGecko.GetBaseURL() + path

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	key, val := c.cfg.CoinGecko.GetAuthHeader()
	req.Header.Set(key, val)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}
