package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/coingecko/coingecko-cli/internal/provider/coingecko"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkets_DryRun(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not make HTTP call in dry-run mode")
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "markets", "--total", "300", "--dry-run", "-o", "json")
	require.NoError(t, err)

	var out dryRunOutput
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	assert.Equal(t, "GET", out.Method)
	assert.Equal(t, "usd", out.Params["vs_currency"])
	assert.Contains(t, out.URL, "/coins/markets")
	require.NotNil(t, out.Pagination)
	assert.Equal(t, 300, out.Pagination.TotalRequested)
	assert.Equal(t, 250, out.Pagination.PerPage)
	assert.Equal(t, 2, out.Pagination.Pages)
}

func TestMarkets_JSONOutput(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/coins/markets", r.URL.Path)
		assert.Equal(t, "usd", r.URL.Query().Get("vs_currency"))
		assert.Equal(t, "market_cap_desc", r.URL.Query().Get("order"))
		assert.Equal(t, "250", r.URL.Query().Get("per_page"))
		coins := []coingecko.MarketCoin{
			{ID: "bitcoin", Name: "Bitcoin", Symbol: "btc", MarketCapRank: 1, CurrentPrice: 50000},
			{ID: "ethereum", Name: "Ethereum", Symbol: "eth", MarketCapRank: 2, CurrentPrice: 3000},
		}
		_ = json.NewEncoder(w).Encode(coins)
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "markets", "--total", "2", "-o", "json")
	require.NoError(t, err)

	var coins []coingecko.MarketCoin
	require.NoError(t, json.Unmarshal([]byte(stdout), &coins))
	assert.Len(t, coins, 2)
	assert.Equal(t, "bitcoin", coins[0].ID)
}

func TestMarkets_CategoryParam(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "layer-1", r.URL.Query().Get("category"))
		_ = json.NewEncoder(w).Encode([]coingecko.MarketCoin{})
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	_, _, err := executeCommand(t, "markets", "--category", "layer-1", "--total", "1", "-o", "json")
	require.NoError(t, err)
}
