package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/coingecko/coingecko-cli/internal/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearch_DryRun(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not make HTTP call in dry-run mode")
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "search", "bitcoin", "--dry-run", "-o", "json")
	require.NoError(t, err)

	var out dryRunOutput
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	assert.Equal(t, "GET", out.Method)
	assert.Equal(t, "bitcoin", out.Params["query"])
	assert.Contains(t, out.URL, "/search")
}

func TestSearch_JSONOutput(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/search", r.URL.Path)
		assert.Equal(t, "btc", r.URL.Query().Get("query"))
		resp := api.SearchResponse{
			Coins: []api.SearchCoin{
				{ID: "bitcoin", Name: "Bitcoin", Symbol: "btc", MarketCapRank: 1},
				{ID: "bitcoin-cash", Name: "Bitcoin Cash", Symbol: "bch", MarketCapRank: 20},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "search", "btc", "--limit", "1", "-o", "json")
	require.NoError(t, err)

	var coins []api.SearchCoin
	require.NoError(t, json.Unmarshal([]byte(stdout), &coins))
	assert.Len(t, coins, 1)
	assert.Equal(t, "bitcoin", coins[0].ID)
}

func TestSearch_NegativeLimit(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		resp := api.SearchResponse{
			Coins: []api.SearchCoin{
				{ID: "bitcoin", Name: "Bitcoin", Symbol: "btc", MarketCapRank: 1},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "search", "btc", "--limit", "-1", "-o", "json")
	require.NoError(t, err)

	var coins []api.SearchCoin
	require.NoError(t, json.Unmarshal([]byte(stdout), &coins))
	assert.Len(t, coins, 0)
}

func TestSearch_MissingArg(t *testing.T) {
	_, _, err := executeCommand(t, "search")
	require.Error(t, err)
}

func TestSearch_APIError(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = fmt.Fprint(w, `{"status":{"error_code":500,"error_message":"Server error"}}`)
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	_, _, err := executeCommand(t, "search", "btc", "-o", "json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}
