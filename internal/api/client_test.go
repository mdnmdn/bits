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
	cfg := &config.Config{APIKey: "pro-key", Tier: config.TierPro}
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

func TestError401PlanRestricted(t *testing.T) {
	// CoinGecko sometimes returns 401 for plan-restricted endpoints
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"status":{"error_code":401,"error_message":"You need to upgrade to a paid plan to access this endpoint"}}`))
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPlanRestricted)
	assert.Contains(t, err.Error(), "upgrade")
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

func TestRequirePaid(t *testing.T) {
	cfg := &config.Config{Tier: config.TierDemo}
	c := NewClient(cfg)
	assert.ErrorIs(t, c.requirePaid(), ErrPlanRestricted)

	cfg2 := &config.Config{Tier: config.TierPro}
	c2 := NewClient(cfg2)
	assert.NoError(t, c2.requirePaid())
}

func TestIsPlanRestricted(t *testing.T) {
	assert.True(t, isPlanRestricted("You need to upgrade your plan"))
	assert.True(t, isPlanRestricted("Access restricted for this tier"))
	assert.True(t, isPlanRestricted("Please subscribe to use this endpoint"))
	assert.True(t, isPlanRestricted("Permission denied for your plan"))
	assert.False(t, isPlanRestricted("Invalid API key"))
	assert.False(t, isPlanRestricted(""))
}
