// Package bitvavo implements the Bitvavo provider: public spot market data only
// (server time, exchange info, 24h ticker, candles). It has no credentials and
// no account or trading endpoints. Bitvavo markets are already in the canonical
// bits form (BTC-EUR).
//
// Fees: Bitvavo publishes fees only through the authenticated /account/fees
// endpoint, so MakerFee and TakerFee stay unset. Each market reports a
// feeCategory (kept in Symbol.Extra) that maps to the public fee tiers.
package bitvavo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/mdnmdn/bits/capability"
	"github.com/mdnmdn/bits/model"
)

const (
	providerID     = "bitvavo"
	defaultBaseURL = "https://api.bitvavo.com/v2"
)

// Config holds the Bitvavo provider settings. All fields are optional.
type Config struct {
	// BaseURL overrides the API root (default https://api.bitvavo.com/v2).
	BaseURL string
}

// Client is the Bitvavo provider.
type Client struct {
	cfg        Config
	httpClient *http.Client
	userAgent  string
}

// NewClient creates a Bitvavo client.
func NewClient(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	return &Client{cfg: cfg, httpClient: &http.Client{Timeout: 30 * time.Second}}
}

// ID returns the provider identifier.
func (c *Client) ID() string { return providerID }

// SetUserAgent sets the User-Agent header for API requests.
func (c *Client) SetUserAgent(ua string) { c.userAgent = ua }

// Capabilities returns the capability matrix: spot only.
func (c *Client) Capabilities() capability.CapabilityMatrix {
	s := capability.MarketSpot
	return capability.NewCapabilityMatrix(
		capability.CapabilityKey{Market: s, Feature: capability.FeatureServerTime},
		capability.CapabilityKey{Market: s, Feature: capability.FeatureExchangeInfo},
		capability.CapabilityKey{Market: s, Feature: capability.FeatureTicker24h},
		capability.CapabilityKey{Market: s, Feature: capability.FeatureCandles},
	)
}

// get performs a public GET. path starts with "/"; query has no leading "?".
func (c *Client) get(ctx context.Context, path, query string) ([]byte, error) {
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
	return body, nil
}

func providerErr(kind model.ErrorKind, msg string, cause error) *model.ProviderError {
	return &model.ProviderError{Kind: kind, ProviderID: providerID, ProviderMessage: msg, Cause: cause}
}

// httpErr maps an HTTP status to an ErrorKind. Bitvavo error bodies look like
// {"errorCode":205,"error":"market parameter is invalid."} and ride on 400/429.
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
	return &model.ProviderError{
		Kind:            kind,
		ProviderID:      providerID,
		ProviderCode:    fmt.Sprintf("%d", status),
		ProviderMessage: body,
		HTTPStatus:      status,
	}
}

func wrapTransportError(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return providerErr(model.ErrKindCanceled, err.Error(), err)
	}
	var ne net.Error
	if errors.As(err, &ne) {
		return providerErr(model.ErrKindNetwork, err.Error(), err)
	}
	return err
}
