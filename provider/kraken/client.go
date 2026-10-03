// Package kraken implements the Kraken provider (public market data only):
// futures on futures.kraken.com, and spot plus margin on api.kraken.com
// (margin reads the same spot pairs and endpoints). No credentials, no
// trading, no account endpoints, no streams.
package kraken

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mdnmdn/bits/capability"
	"github.com/mdnmdn/bits/model"
)

const (
	providerID         = "kraken"
	defaultBaseURL     = "https://futures.kraken.com"
	defaultSpotBaseURL = "https://api.kraken.com"
)

// Config holds the provider settings. There are no credentials.
type Config struct {
	// BaseURL overrides the REST host (default https://futures.kraken.com).
	BaseURL string
	// SpotBaseURL overrides the spot REST host (default https://api.kraken.com).
	SpotBaseURL string
	// SpotEnabled, MarginEnabled and FuturesEnabled select the markets in
	// Capabilities. When none is set, all three are enabled.
	SpotEnabled, MarginEnabled, FuturesEnabled bool
}

// Client is the Kraken provider.
type Client struct {
	cfg        Config
	httpClient *http.Client
	userAgent  string
}

// NewClient creates a Kraken client.
func NewClient(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	if cfg.SpotBaseURL == "" {
		cfg.SpotBaseURL = defaultSpotBaseURL
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	cfg.SpotBaseURL = strings.TrimRight(cfg.SpotBaseURL, "/")
	return &Client{cfg: cfg, httpClient: &http.Client{Timeout: 30 * time.Second}}
}

// ID returns the provider identifier.
func (c *Client) ID() string { return providerID }

// SetUserAgent sets the User-Agent header for API requests.
func (c *Client) SetUserAgent(ua string) { c.userAgent = ua }

// Capabilities returns the capability matrix. Spot and margin share the spot
// REST endpoints; futures has funding rates on top. No streams.
func (c *Client) Capabilities() capability.CapabilityMatrix {
	spot, margin, fut := c.cfg.SpotEnabled, c.cfg.MarginEnabled, c.cfg.FuturesEnabled
	if !spot && !margin && !fut {
		spot, margin, fut = true, true, true
	}
	m := capability.CapabilityMatrix{}
	add := func(mkt capability.MarketType, fs ...capability.Feature) {
		for _, f := range fs {
			m[capability.CapabilityKey{Market: mkt, Feature: f}] = true
		}
	}
	common := []capability.Feature{
		capability.FeatureServerTime, capability.FeatureExchangeInfo, capability.FeaturePrice,
		capability.FeatureCandles, capability.FeatureTicker24h, capability.FeatureOrderBook,
	}
	if spot {
		add(capability.MarketSpot, common...)
	}
	if margin {
		add(capability.MarketMargin, common...)
	}
	if fut {
		add(capability.MarketFutures, common...)
		add(capability.MarketFutures, capability.FeatureFundingRates)
	}
	return m
}

// get performs a GET on path (with its query string) and decodes the JSON
// body into out. Derivatives API bodies carry result:"success" or
// result:"error" with an error string; the charts API has no result field.
func (c *Client) get(ctx context.Context, path string, out any) error {
	body, err := c.fetch(ctx, c.cfg.BaseURL, path)
	if err != nil {
		return err
	}
	var status struct {
		Result string `json:"result"`
		Error  string `json:"error"`
	}
	if json.Unmarshal(body, &status) == nil && status.Result == "error" {
		return apiErr(status.Error)
	}
	if err := json.Unmarshal(body, out); err != nil {
		return providerErr(model.ErrKindParse, "parse response: "+err.Error(), err)
	}
	return nil
}

// fetch performs a GET on base+path and returns the body of a 2xx response.
func (c *Client) fetch(ctx context.Context, base, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+path, nil)
	if err != nil {
		return nil, providerErr(model.ErrKindInvalidRequest, err.Error(), err)
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, providerErr(model.ErrKindCanceled, err.Error(), err)
		}
		var ne net.Error
		if errors.As(err, &ne) {
			return nil, providerErr(model.ErrKindNetwork, err.Error(), err)
		}
		return nil, providerErr(model.ErrKindNetwork, "request failed: "+err.Error(), err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, providerErr(model.ErrKindNetwork, "read response: "+err.Error(), err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, httpErr(resp.StatusCode, string(body))
	}
	return body, nil
}

func providerErr(kind model.ErrorKind, msg string, cause error) *model.ProviderError {
	return &model.ProviderError{Kind: kind, ProviderID: providerID, ProviderMessage: msg, Cause: cause}
}

func httpErr(status int, body string) *model.ProviderError {
	kind := model.ErrKindUnknown
	switch {
	case status == 401 || status == 403:
		kind = model.ErrKindAuth
	case status == 404:
		kind = model.ErrKindNotFound
	case status == 429:
		kind = model.ErrKindRateLimit
	case status == 400:
		kind = model.ErrKindInvalidRequest
	case status >= 500:
		kind = model.ErrKindServerError
	}
	return &model.ProviderError{Kind: kind, ProviderID: providerID, ProviderCode: strconv.Itoa(status), ProviderMessage: body, HTTPStatus: status}
}

func apiErr(code string) *model.ProviderError {
	kind := model.ErrKindUnknown
	switch code {
	case "apiLimitExceeded":
		kind = model.ErrKindRateLimit
	case "authenticationError":
		kind = model.ErrKindAuth
	case "invalidArgument", "unavailable":
		kind = model.ErrKindInvalidRequest
	case "marketUnavailable":
		kind = model.ErrKindNotFound
	}
	return &model.ProviderError{Kind: kind, ProviderID: providerID, ProviderCode: code, ProviderMessage: code}
}
