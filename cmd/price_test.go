package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/coingecko/coingecko-cli/internal/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrice_MissingIDsAndSymbols(t *testing.T) {
	_, _, err := executeCommand(t, "price", "-o", "json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--ids or --symbols")
}

func TestPrice_DryRun(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not make HTTP call in dry-run mode")
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "price", "--ids", "bitcoin", "--dry-run", "-o", "json")
	require.NoError(t, err)

	var out dryRunOutput
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	assert.Equal(t, "GET", out.Method)
	assert.Equal(t, "bitcoin", out.Params["ids"])
	assert.Equal(t, "usd", out.Params["vs_currencies"])
	assert.Contains(t, out.URL, "/simple/price")
}

func TestPrice_ByIDs_JSONOutput(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/simple/price", r.URL.Path)
		assert.Equal(t, "bitcoin", r.URL.Query().Get("ids"))
		resp := api.PriceResponse{
			"bitcoin": {"usd": 50000, "usd_24h_change": 2.5},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "price", "--ids", "bitcoin", "-o", "json")
	require.NoError(t, err)

	var prices api.PriceResponse
	require.NoError(t, json.Unmarshal([]byte(stdout), &prices))
	assert.Equal(t, float64(50000), prices["bitcoin"]["usd"])
}

func TestPrice_BySymbols_JSONOutput(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/search" {
			// Verify the search query matches the requested symbol.
			assert.Equal(t, "btc", r.URL.Query().Get("query"))
			resp := api.SearchResponse{
				Coins: []api.SearchCoin{
					{ID: "bitcoin", Symbol: "btc", MarketCapRank: 1},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		// Verify the resolved ID is what gets requested.
		assert.Equal(t, "/simple/price", r.URL.Path)
		assert.Equal(t, "bitcoin", r.URL.Query().Get("ids"))
		assert.Equal(t, "usd", r.URL.Query().Get("vs_currencies"))
		resp := api.PriceResponse{
			"bitcoin": {"usd": 50000},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "price", "--symbols", "btc", "-o", "json")
	require.NoError(t, err)

	var prices api.PriceResponse
	require.NoError(t, json.Unmarshal([]byte(stdout), &prices))
	assert.Contains(t, prices, "bitcoin")
}

func TestPrice_SymbolNoMatch(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/search" {
			// Return coins that don't match the symbol
			resp := api.SearchResponse{
				Coins: []api.SearchCoin{
					{ID: "something", Symbol: "xyz", MarketCapRank: 100},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		t.Fatal("should not call price API when no symbols resolved")
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	_, stderr, err := executeCommand(t, "price", "--symbols", "notarealcoin", "-o", "json")
	require.Error(t, err)
	assert.Contains(t, stderr, "no exact match")
	assert.Contains(t, err.Error(), "no valid coins found")
}
