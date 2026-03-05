package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"coingecko-cli/internal/config"

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

func TestError401(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	})
	defer srv.Close()

	var result map[string]any
	err := c.get(context.Background(), "/test", &result)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidAPIKey)
}

func TestError403(t *testing.T) {
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

func TestRequirePaid(t *testing.T) {
	cfg := &config.Config{Tier: config.TierDemo}
	c := NewClient(cfg)
	assert.ErrorIs(t, c.requirePaid(), ErrPlanRestricted)

	cfg2 := &config.Config{Tier: config.TierPro}
	c2 := NewClient(cfg2)
	assert.NoError(t, c2.requirePaid())
}
