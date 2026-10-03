// Package kraken implements the Kraken Futures provider (public market data
// only: perpetual and dated futures on futures.kraken.com). Spot Kraken is not
// covered. No credentials, no trading, no account endpoints.
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
	providerID     = "kraken"
	defaultBaseURL = "https://futures.kraken.com"
)

// Config holds the provider settings. There are no credentials.
type Config struct {
	// BaseURL overrides the REST host (default https://futures.kraken.com).
	BaseURL string
}

// Client is the Kraken Futures provider.
type Client struct {
	cfg        Config
	httpClient *http.Client
	userAgent  string
}

// NewClient creates a Kraken Futures client.
func NewClient(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	return &Client{cfg: cfg, httpClient: &http.Client{Timeout: 30 * time.Second}}
}

// ID returns the provider identifier.
func (c *Client) ID() string { return providerID }

// SetUserAgent sets the User-Agent header for API requests.
func (c *Client) SetUserAgent(ua string) { c.userAgent = ua }

// Capabilities returns the capability matrix: futures only.
func (c *Client) Capabilities() capability.CapabilityMatrix {
	f := capability.MarketFutures
	return capability.NewCapabilityMatrix(
		capability.CapabilityKey{Market: f, Feature: capability.FeatureServerTime},
		capability.CapabilityKey{Market: f, Feature: capability.FeatureExchangeInfo},
		capability.CapabilityKey{Market: f, Feature: capability.FeatureTicker24h},
		capability.CapabilityKey{Market: f, Feature: capability.FeatureCandles},
	)
}

// get performs a GET on path (with its query string) and decodes the JSON
// body into out. Derivatives API bodies carry result:"success" or
// result:"error" with an error string; the charts API has no result field.
func (c *Client) get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.cfg.BaseURL+path, nil)
	if err != nil {
		return providerErr(model.ErrKindInvalidRequest, err.Error(), err)
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return providerErr(model.ErrKindCanceled, err.Error(), err)
		}
		var ne net.Error
		if errors.As(err, &ne) {
			return providerErr(model.ErrKindNetwork, err.Error(), err)
		}
		return providerErr(model.ErrKindNetwork, "request failed: "+err.Error(), err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return providerErr(model.ErrKindNetwork, "read response: "+err.Error(), err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return httpErr(resp.StatusCode, string(body))
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
