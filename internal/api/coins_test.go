package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/coingecko/coingecko-cli/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchAllMarkets_SinglePage(t *testing.T) {
	var calls int32
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		assert.Equal(t, "250", r.URL.Query().Get("per_page"))
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		coins := make([]MarketCoin, 10)
		for i := range coins {
			coins[i] = MarketCoin{ID: fmt.Sprintf("coin-%d", i), MarketCapRank: i + 1}
		}
		_ = json.NewEncoder(w).Encode(coins)
	})
	defer srv.Close()

	result, err := c.FetchAllMarkets(context.Background(), "usd", 10, "market_cap_desc", "")
	require.NoError(t, err)
	assert.Len(t, result, 10)
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls))
}

func TestFetchAllMarkets_MultiPage(t *testing.T) {
	var calls int32
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		atomic.AddInt32(&calls, 1)
		// Return 250 coins per page (full page)
		coins := make([]MarketCoin, 250)
		for i := range coins {
			pageNum := 1
			if page == "2" {
				pageNum = 2
			}
			coins[i] = MarketCoin{
				ID:            fmt.Sprintf("coin-p%s-%d", page, i),
				MarketCapRank: (pageNum-1)*250 + i + 1,
			}
		}
		_ = json.NewEncoder(w).Encode(coins)
	})
	defer srv.Close()

	result, err := c.FetchAllMarkets(context.Background(), "usd", 300, "market_cap_desc", "")
	require.NoError(t, err)
	// Requested 300, should trim from 500 (2 pages × 250)
	assert.Len(t, result, 300)
	assert.Equal(t, int32(2), atomic.LoadInt32(&calls))
	// First coin from page 1, last from page 2
	assert.Equal(t, "coin-p1-0", result[0].ID)
	assert.Equal(t, "coin-p2-49", result[299].ID)
}

func TestFetchAllMarkets_PartialLastPage(t *testing.T) {
	var calls int32
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		var count int
		if n == 1 {
			count = 250 // full first page
		} else {
			count = 100 // partial second page — signals end of data
		}
		coins := make([]MarketCoin, count)
		for i := range coins {
			coins[i] = MarketCoin{ID: fmt.Sprintf("coin-%d-%d", n, i)}
		}
		_ = json.NewEncoder(w).Encode(coins)
	})
	defer srv.Close()

	// Request 500 but only 350 exist
	result, err := c.FetchAllMarkets(context.Background(), "usd", 500, "market_cap_desc", "")
	require.NoError(t, err)
	assert.Len(t, result, 350)
	assert.Equal(t, int32(2), atomic.LoadInt32(&calls)) // stops after partial page
}

func TestFetchAllMarkets_APIErrorMidPagination(t *testing.T) {
	var calls int32
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n == 2 {
			w.WriteHeader(500)
			fmt.Fprint(w, `{"status":{"error_code":500,"error_message":"Internal error"}}`)
			return
		}
		coins := make([]MarketCoin, 250)
		for i := range coins {
			coins[i] = MarketCoin{ID: fmt.Sprintf("coin-%d", i)}
		}
		_ = json.NewEncoder(w).Encode(coins)
	})
	defer srv.Close()

	_, err := c.FetchAllMarkets(context.Background(), "usd", 500, "market_cap_desc", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Equal(t, int32(2), atomic.LoadInt32(&calls))
}

func TestFetchAllMarkets_PassesQueryParams(t *testing.T) {
	var gotParams []map[string]string
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		gotParams = append(gotParams, map[string]string{
			"vs_currency": q.Get("vs_currency"),
			"order":       q.Get("order"),
			"category":    q.Get("category"),
			"per_page":    q.Get("per_page"),
			"page":        q.Get("page"),
		})
		_ = json.NewEncoder(w).Encode([]MarketCoin{})
	})
	defer srv.Close()

	_, _ = c.FetchAllMarkets(context.Background(), "eur", 10, "volume_desc", "layer-1")
	require.Len(t, gotParams, 1)
	assert.Equal(t, "eur", gotParams[0]["vs_currency"])
	assert.Equal(t, "volume_desc", gotParams[0]["order"])
	assert.Equal(t, "layer-1", gotParams[0]["category"])
	assert.Equal(t, "250", gotParams[0]["per_page"])
	assert.Equal(t, "1", gotParams[0]["page"])
}

func TestFetchAllMarkets_EmptyCategory(t *testing.T) {
	c, srv := testClient(func(w http.ResponseWriter, r *http.Request) {
		// When category is "", the param should not be set
		assert.Empty(t, r.URL.Query().Get("category"))
		_ = json.NewEncoder(w).Encode([]MarketCoin{})
	})
	defer srv.Close()

	_, err := c.FetchAllMarkets(context.Background(), "usd", 10, "market_cap_desc", "")
	require.NoError(t, err)
}

func TestFetchAllMarkets_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			cancel() // cancel after first request
		}
		coins := make([]MarketCoin, 250)
		for i := range coins {
			coins[i] = MarketCoin{ID: fmt.Sprintf("coin-%d", i)}
		}
		_ = json.NewEncoder(w).Encode(coins)
	}))
	defer srv.Close()

	cfg := &config.Config{APIKey: "test-key", Tier: config.TierDemo}
	c := NewClient(cfg)
	c.SetBaseURL(srv.URL)

	_, err := c.FetchAllMarkets(ctx, "usd", 500, "market_cap_desc", "")
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}
