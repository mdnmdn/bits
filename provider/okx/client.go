// Package okx implements the OKX provider (public market data only: spot and
// perpetual swaps). EEA customers must use https://eea.okx.com, the default.
package okx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mdnmdn/bits/capability"
	"github.com/mdnmdn/bits/config"
	"github.com/mdnmdn/bits/model"
)

const (
	providerID = "okx"
	// defaultBaseURL is the EEA host. www.okx.com serves the same public data.
	defaultBaseURL = "https://eea.okx.com"
	// defaultPageDelay keeps paged history requests under the 20 req / 2 s limit
	// of /market/history-candles.
	defaultPageDelay = 110 * time.Millisecond
)

// Client is the OKX provider. It has no credentials: public endpoints only.
type Client struct {
	cfg        config.OKXConfig
	httpClient *http.Client
	userAgent  string
	pageDelay  time.Duration
}

// NewClient creates an OKX client from the given config.
func NewClient(cfg config.OKXConfig) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		pageDelay:  defaultPageDelay,
	}
}

// ID returns the provider identifier.
func (c *Client) ID() string { return providerID }

// SetUserAgent sets the User-Agent header for API requests.
func (c *Client) SetUserAgent(ua string) { c.userAgent = ua }

// Capabilities returns the capability matrix based on the configured markets.
// Spot is enabled if nothing else is enabled. Futures means perpetual swaps.
func (c *Client) Capabilities() capability.CapabilityMatrix {
	s := capability.MarketSpot
	f := capability.MarketFutures

	spotEnabled := c.cfg.IsSpotEnabled()
	futuresEnabled := c.cfg.IsFuturesEnabled()
	if !spotEnabled && !futuresEnabled {
		spotEnabled = true
	}

	matrix := capability.CapabilityMatrix{}
	matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureServerTime}] = true
	if spotEnabled {
		matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureExchangeInfo}] = true
		matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureCandles}] = true
		matrix[capability.CapabilityKey{Market: s, Feature: capability.FeatureTicker24h}] = true
	}
	if futuresEnabled {
		matrix[capability.CapabilityKey{Market: f, Feature: capability.FeatureExchangeInfo}] = true
		matrix[capability.CapabilityKey{Market: f, Feature: capability.FeatureCandles}] = true
		matrix[capability.CapabilityKey{Market: f, Feature: capability.FeatureTicker24h}] = true
	}
	return matrix
}

// okxEnvelope is the common OKX response wrapper.
type okxEnvelope struct {
	Code string          `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

// get performs a public GET and returns the raw "data" array after checking
// the envelope code. query is the raw query string without "?".
func (c *Client) get(ctx context.Context, path, query string) (json.RawMessage, error) {
	url := c.cfg.BaseURL + path
	if query != "" {
		url += "?" + query
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, wrapTransportError(err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, httpErr(resp.StatusCode, string(body))
	}

	var env okxEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, providerErr(model.ErrKindParse, "failed to parse response", err)
	}
	if env.Code != "0" {
		return nil, apiErr(env.Code, env.Msg)
	}
	return env.Data, nil
}
