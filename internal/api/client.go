package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/coingecko/coingecko-cli/internal/config"
)

const (
	maxErrorBodySize    = 1 << 20 // 1MB
	maxResponseBodySize = 50 << 20 // 50MB — guards against pathological upstream responses
)

var (
	ErrInvalidAPIKey  = fmt.Errorf("invalid API key — check your key with `cg status` or set a new one with `cg auth`")
	ErrPlanRestricted = fmt.Errorf("this endpoint requires a paid plan — upgrade at https://www.coingecko.com/en/api/pricing")
	ErrRateLimited    = fmt.Errorf("rate limited — please wait and try again")
)

// apiErrorResponse covers CoinGecko's error JSON formats.
type apiErrorResponse struct {
	Status *struct {
		ErrorCode    int    `json:"error_code"`
		ErrorMessage string `json:"error_message"`
	} `json:"status"`
	Error string `json:"error"`
}

// extractMessage returns the best error message from the response body.
func (e *apiErrorResponse) extractMessage() string {
	if e.Status != nil && e.Status.ErrorMessage != "" {
		return e.Status.ErrorMessage
	}
	if e.Error != "" {
		return e.Error
	}
	return ""
}

type Client struct {
	http    *http.Client
	baseURL string
	cfg     *config.Config
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		http:    &http.Client{Timeout: 30 * time.Second},
		baseURL: cfg.BaseURL(),
		cfg:     cfg,
	}
}

func NewClientWithHTTP(cfg *config.Config, httpClient *http.Client) *Client {
	return &Client{
		http:    httpClient,
		baseURL: cfg.BaseURL(),
		cfg:     cfg,
	}
}

func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

func (c *Client) get(ctx context.Context, path string, result any) error {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	c.cfg.ApplyAuth(req)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.handleError(resp)
	}

	lr := io.LimitReader(resp.Body, maxResponseBodySize+1)
	dec := json.NewDecoder(lr)
	if err := dec.Decode(result); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}
	return nil
}

func (c *Client) handleError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySize))

	// Parse CoinGecko error body for better classification.
	var apiErr apiErrorResponse
	json.Unmarshal(body, &apiErr)
	msg := apiErr.extractMessage()

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		if msg != "" {
			return fmt.Errorf("%w: %s", ErrInvalidAPIKey, msg)
		}
		return ErrInvalidAPIKey

	case http.StatusForbidden:
		if msg != "" {
			return fmt.Errorf("%w: %s", ErrPlanRestricted, msg)
		}
		return ErrPlanRestricted

	case http.StatusTooManyRequests:
		if retry := resp.Header.Get("Retry-After"); retry != "" {
			if secs, err := strconv.Atoi(retry); err == nil && secs > 0 {
				return fmt.Errorf("rate limited — retry after %d seconds", secs)
			}
			// Try HTTP-date format (RFC1123).
			if t, err := time.Parse(time.RFC1123, retry); err == nil {
				if wait := time.Until(t).Seconds(); wait > 0 {
					return fmt.Errorf("rate limited — retry after %.0f seconds", wait)
				}
			}
		}
		return ErrRateLimited

	default:
		if msg != "" {
			return fmt.Errorf("API error %d: %s", resp.StatusCode, msg)
		}
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
}

func (c *Client) requirePaid() error {
	if !c.cfg.IsPaid() {
		return ErrPlanRestricted
	}
	return nil
}
