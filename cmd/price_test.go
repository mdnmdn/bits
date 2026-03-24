package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/coingecko/coingecko-cli/internal/model"

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

func TestPrice_DryRun_Symbols(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not make HTTP call in dry-run mode")
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "price", "--symbols", "btc", "--dry-run", "-o", "json")
	require.NoError(t, err)

	var out dryRunOutput
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	assert.Equal(t, "btc", out.Params["symbols"])
	assert.Contains(t, out.URL, "/simple/price")
}

func TestPrice_ByIDs_JSONOutput(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/simple/price", r.URL.Path)
		assert.Equal(t, "bitcoin", r.URL.Query().Get("ids"))
		resp := model.PriceResponse{
			"bitcoin": {"usd": 50000, "usd_24h_change": 2.5},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "price", "--ids", "bitcoin", "-o", "json")
	require.NoError(t, err)

	var prices model.PriceResponse
	require.NoError(t, json.Unmarshal([]byte(stdout), &prices))
	assert.Equal(t, float64(50000), prices["bitcoin"]["usd"])
}

func TestPrice_BySymbols_JSONOutput(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		// Symbols are passed directly — no /search call.
		assert.Equal(t, "/simple/price", r.URL.Path)
		assert.Equal(t, "btc", r.URL.Query().Get("symbols"))
		assert.Equal(t, "usd", r.URL.Query().Get("vs_currencies"))
		resp := model.PriceResponse{
			"bitcoin": {"usd": 50000},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	stdout, _, err := executeCommand(t, "price", "--symbols", "btc", "-o", "json")
	require.NoError(t, err)

	var prices model.PriceResponse
	require.NoError(t, json.Unmarshal([]byte(stdout), &prices))
	assert.Contains(t, prices, "bitcoin")
}

func TestPrice_SymbolNoResults(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		// API returns empty response for unknown symbol.
		_ = json.NewEncoder(w).Encode(model.PriceResponse{})
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	_, _, err := executeCommand(t, "price", "--symbols", "notarealcoin", "-o", "json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no valid coins found")
}

func TestPrice_BySymbols_TableOutput_NoFalseWarning(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		// API returns coin ID as key, not the requested symbol.
		resp := model.PriceResponse{
			"bitcoin": {"usd": 50000, "usd_24h_change": 1.0},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	_, stderr, err := executeCommand(t, "price", "--symbols", "btc")
	require.NoError(t, err)
	assert.NotContains(t, stderr, "no data returned")
}

func TestPrice_PartialMiss_WarnsOnStderr(t *testing.T) {
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		// Only return data for bitcoin, not for "missing".
		resp := model.PriceResponse{
			"bitcoin": {"usd": 50000, "usd_24h_change": 1.0},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()
	withTestClientDemo(t, srv)

	_, stderr, err := executeCommand(t, "price", "--ids", "bitcoin,missing")
	require.NoError(t, err)
	assert.Contains(t, stderr, `no data returned for "missing"`)
}
