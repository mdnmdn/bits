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

// RateLimitError carries the Retry-After metadata from a 429 response.
type RateLimitError struct {
	RetryAfter int // seconds; 0 if Retry-After header was absent
}

func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("rate limited — retry after %d seconds", e.RetryAfter)
	}
	return "rate limited — please wait and try again"
}

func (e *RateLimitError) Is(target error) bool {
	return target == ErrRateLimited
}

// apiErrorResponse covers CoinGecko's error JSON formats.
// The API may return:
//
//	{"status": {"error_code": N, "error_message": "..."}}
//	{"error": {"status": {"error_code": N, "error_message": "..."}}}
//	{"error": "some string"}
type apiErrorResponse struct {
	Status *struct {
		ErrorCode    int    `json:"error_code"`
		ErrorMessage string `json:"error_message"`
	} `json:"status"`
	Error json.RawMessage `json:"error"`
}

// extractMessage returns the best error message from the response body.
func (e *apiErrorResponse) extractMessage() string {
	if e.Status != nil && e.Status.ErrorMessage != "" {
		return e.Status.ErrorMessage
	}
	if len(e.Error) > 0 {
		// Try nested {"error": {"status": {"error_message": "..."}}}
		var nested struct {
			Status *struct {
				ErrorMessage string `json:"error_message"`
			} `json:"status"`
		}
		if json.Unmarshal(e.Error, &nested) == nil && nested.Status != nil && nested.Status.ErrorMessage != "" {
			return nested.Status.ErrorMessage
		}
		// Try simple {"error": "string"}
		var s string
		if json.Unmarshal(e.Error, &s) == nil && s != "" {
			return s
		}
	}
	return ""
}

type Client struct {
	http       *http.Client
	baseURLVal string // override; empty = use cfg.BaseURL()
	cfg        *config.Config
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		http: &http.Client{Timeout: 30 * time.Second},
		cfg:  cfg,
	}
}

func NewClientWithHTTP(cfg *config.Config, httpClient *http.Client) *Client {
	return &Client{
		http: httpClient,
		cfg:  cfg,
	}
}

func (c *Client) SetBaseURL(url string) {
	c.baseURLVal = url
}

func (c *Client) baseURL() string {
	if c.baseURLVal != "" {
		return c.baseURLVal
	}
	return c.cfg.BaseURL()
}

func (c *Client) get(ctx context.Context, path string, result any) error {
	url := c.baseURL() + path

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
			return fmt.Errorf("%s (%w)", msg, ErrInvalidAPIKey)
		}
		return ErrInvalidAPIKey

	case http.StatusForbidden:
		if msg != "" {
			return fmt.Errorf("%s (%w)", msg, ErrPlanRestricted)
		}
		return ErrPlanRestricted

	case http.StatusTooManyRequests:
		retryAfter := 0
		if retry := resp.Header.Get("Retry-After"); retry != "" {
			if secs, err := strconv.Atoi(retry); err == nil && secs > 0 {
				retryAfter = secs
			}
		}
		return &RateLimitError{RetryAfter: retryAfter}

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
