package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coingecko/coingecko-cli/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testClient(handler http.HandlerFunc) (*Client, *httptest.Server) {
	srv := httptest.NewServer(handler)
	cfg := &config.Config{APIKey: "test-key", Tier: config.TierDemo}
	c := NewClient(cfg)
	c.SetBaseURL(srv.URL)
	return c, srv
}

func TestAuthHeadersSent(t *testing.T) {
	var gotHeader string
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("x-cg-demo-api-key")
		w.WriteHeader(200)
		w.Write([]byte("{}"))
	})
	defer srv.Close()

	var result map[string]any
	c.get(context.Background(), "/test", &result)
	assert.Equal(t, "test-key", gotHeader)
}

func TestProAuthHeaders(t *testing.T) {
	var gotHeader string
	cfg := &config.Config{APIKey: "pro-key", Tier: config.TierPaid}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("x-cg-pro-api-key")
		w.WriteHeader(200)
		w.Write([]byte("{}"))
	}))
	defer srv.Close()

	c := NewClient(cfg)
	c.SetBaseURL(srv.URL)
	var result map[string]any
	c.get(context.Background(), "/test", &result)
	assert.Equal(t, "pro-key", gotHeader)
}

func TestError401InvalidKey(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"status":{"error_code":401,"error_message":"Invalid API key"}}`))
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidAPIKey)
	assert.Contains(t, err.Error(), "Invalid API key")
}

func TestError401EmptyBody(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidAPIKey)
}

func TestError401WithAPIMessage(t *testing.T) {
	// 401 always maps to ErrInvalidAPIKey — the API message is passed through
	// so the user sees the real reason (which may mention plan restrictions)
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"status":{"error_code":401,"error_message":"You need to upgrade to a paid plan"}}`))
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidAPIKey)
	assert.Contains(t, err.Error(), "You need to upgrade to a paid plan")
}

func TestError401NestedErrorFormat(t *testing.T) {
	// CoinGecko sometimes wraps errors as {"error":{"status":{"error_message":"..."}}}
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"error":{"status":{"timestamp":"2026-03-05T08:16:12.383+00:00","error_code":10012,"error_message":"Your request exceeds the allowed time range."}}}`))
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidAPIKey)
	assert.Contains(t, err.Error(), "allowed time range")
}

func TestError403(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		w.Write([]byte(`{"error":"Your plan does not have access to this endpoint"}`))
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPlanRestricted)
	assert.Contains(t, err.Error(), "access")
}

func TestError403EmptyBody(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPlanRestricted)
}

func TestError429(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(429)
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "retry after 30 seconds")
}

func TestError429NoRetryAfter(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRateLimited)
}

func TestErrorUnknownStatusWithBody(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"status":{"error_code":500,"error_message":"Internal server error"}}`))
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API error 500")
	assert.Contains(t, err.Error(), "Internal server error")
}

func TestErrorUnknownStatusRawBody(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(502)
		w.Write([]byte("Bad Gateway"))
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API error 502")
	assert.Contains(t, err.Error(), "Bad Gateway")
}

func TestSuccessResponseDecodes(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"bitcoin":{"usd":50000}}`))
	})
	defer srv.Close()

	var result PriceResponse
	err := c.get(context.Background(), "/test", &result)
	require.NoError(t, err)
	assert.Equal(t, float64(50000), result["bitcoin"]["usd"])
}

func TestRetryAfterInvalidFallback(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "invalid-value")
		w.WriteHeader(429)
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRateLimited)
}

func TestRequirePaid(t *testing.T) {
	cfg := &config.Config{Tier: config.TierDemo}
	c := NewClient(cfg)
	assert.ErrorIs(t, c.requirePaid(), ErrPlanRestricted)

	cfg2 := &config.Config{Tier: config.TierPaid}
	c2 := NewClient(cfg2)
	assert.NoError(t, c2.requirePaid())
}

