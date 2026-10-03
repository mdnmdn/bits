// Package bybiteu implements the Bybit EU provider (public spot market data
// only, V5 REST API). EEA customers use https://api.bybit.eu, the default; the
// base URL is configurable. No credentials, no trading, no account endpoints.
package bybiteu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/mdnmdn/bits/capability"
	"github.com/mdnmdn/bits/model"
)

const (
	providerID = "bybiteu"
	// defaultBaseURL is the EEA host named by the Bybit V5 integration guide
	// ("EEA users: use https://api.bybit.eu for mainnet"). The data endpoints
	// are the same as on api.bybit.com.
	defaultBaseURL = "https://api.bybit.eu"
)

// Config holds the provider settings. There are no credentials.
type Config struct {
	// BaseURL overrides the REST host (default https://api.bybit.eu).
	BaseURL string
}

// Client is the Bybit EU provider.
type Client struct {
	cfg        Config
	httpClient *http.Client
	userAgent  string
}

// NewClient creates a Bybit EU client.
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

// envelope is the common V5 response wrapper.
type envelope struct {
	RetCode int             `json:"retCode"`
	RetMsg  string          `json:"retMsg"`
	Result  json.RawMessage `json:"result"`
	Time    int64           `json:"time"`
}

// get performs a GET and returns the decoded envelope.
func (c *Client) get(ctx context.Context, path string, q url.Values) (*envelope, error) {
	u := c.cfg.BaseURL + path
	if len(q) > 0 {
		u += "?" + q.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
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
	var env envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, providerErr(model.ErrKindParse, "parse response: "+err.Error(), err)
	}
	if env.RetCode != 0 {
		return nil, apiErr(env.RetCode, env.RetMsg)
	}
	return &env, nil
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

func apiErr(code int, msg string) *model.ProviderError {
	kind := model.ErrKindUnknown
	switch code {
	case 10006, 10018:
		kind = model.ErrKindRateLimit
	case 10003, 10004, 10005:
		kind = model.ErrKindAuth
	case 10001:
		kind = model.ErrKindInvalidRequest
	}
	return &model.ProviderError{Kind: kind, ProviderID: providerID, ProviderCode: fmt.Sprintf("%d", code), ProviderMessage: msg}
}
