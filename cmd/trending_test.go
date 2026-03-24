package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/coingecko/coingecko-cli/internal/provider/coingecko"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrending_DryRun(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not make HTTP call in dry-run mode")
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "trending", "--dry-run", "-o", "json")
	require.NoError(t, err)

	var out dryRunOutput
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	assert.Contains(t, out.URL, "/search/trending")
}

func TestTrending_ShowMaxDemoRejected(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not make HTTP call when plan-gated")
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	_, _, err := executeCommand(t, "trending", "--show-max", "coins", "-o", "json")
	require.Error(t, err)
	assert.ErrorIs(t, err, coingecko.ErrPlanRestricted)
}

func TestTrending_JSONOutput(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/search/trending", r.URL.Path)
		// Demo tier without --show-max should not send show_max param.
		assert.Empty(t, r.URL.Query().Get("show_max"))
		resp := coingecko.TrendingResponse{
			Coins: []coingecko.TrendingCoinWrapper{
				{Item: coingecko.TrendingCoin{ID: "bitcoin", Name: "Bitcoin", Symbol: "BTC", MarketCapRank: 1}},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "trending", "-o", "json")
	require.NoError(t, err)

	var resp coingecko.TrendingResponse
	require.NoError(t, json.Unmarshal([]byte(stdout), &resp))
	assert.Len(t, resp.Coins, 1)
	assert.Equal(t, "bitcoin", resp.Coins[0].Item.ID)
}

func TestTrending_ShowMaxPaidAllowed(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "coins", r.URL.Query().Get("show_max"))
		resp := coingecko.TrendingResponse{}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()
	withTestClientPaid(t, srv)

	_, _, err := executeCommand(t, "trending", "--show-max", "coins", "-o", "json")
	require.NoError(t, err)
}
