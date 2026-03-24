package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/coingecko/coingecko-cli/internal/model"
	"github.com/coingecko/coingecko-cli/internal/config"
	"github.com/coingecko/coingecko-cli/internal/ws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeStreamer implements the Streamer interface for testing.
type fakeStreamer struct {
	updates   []*ws.CoinUpdate
	connectFn func(ctx context.Context) (<-chan *ws.CoinUpdate, error)
}

func (f *fakeStreamer) Connect(ctx context.Context) (<-chan *ws.CoinUpdate, error) {
	if f.connectFn != nil {
		return f.connectFn(ctx)
	}
	ch := make(chan *ws.CoinUpdate, len(f.updates))
	for _, u := range f.updates {
		ch <- u
	}
	close(ch)
	return ch, nil
}

func (f *fakeStreamer) Close() error { return nil }

func withFakeStreamer(t *testing.T, s Streamer) {
	t.Helper()
	orig := newStreamer
	newStreamer = func(cfg *config.Config, coinIDs []string) Streamer {
		return s
	}
	t.Cleanup(func() { newStreamer = orig })
}

func TestWatch_MissingFlags(t *testing.T) {
	_, _, err := executeCommand(t, "watch", "-o", "json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--ids or --symbols")
}

func TestWatch_UnresolvedSymbol(t *testing.T) {
	// Search returns no matching coins for the symbol.
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/search" {
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{"coins":[]}`))
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{}`))
	})
	defer srv.Close()
	withTestClient(t, srv, "paid")

	_, _, err := executeCommand(t, "watch", "--symbols", "zzzznotacoin", "-o", "json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "none of the provided coins could be found")
}

func TestWatch_SymbolRankDisambiguation(t *testing.T) {
	// Search returns two coins with the same symbol; should pick highest rank.
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/search" {
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{"coins":[
				{"id":"fake-btc","name":"Fake BTC","symbol":"btc","market_cap_rank":999},
				{"id":"bitcoin","name":"Bitcoin","symbol":"BTC","market_cap_rank":1}
			]}`))
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{}`))
	})
	defer srv.Close()
	withTestClient(t, srv, "paid")

	withFakeStreamer(t, &fakeStreamer{
		updates: []*ws.CoinUpdate{
			{CoinID: "bitcoin", Price: 45000},
		},
	})

	stdout, _, err := executeCommand(t, "watch", "--symbols", "btc", "-o", "json")
	require.NoError(t, err)

	lines := splitNonEmpty(stdout)
	require.Len(t, lines, 1)
	var u ws.CoinUpdate
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &u))
	assert.Equal(t, "bitcoin", u.CoinID)
}

// simplePriceServer returns a test server that responds to /simple/price
// with valid price data for the requested coin IDs.
func simplePriceServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		ids := strings.Split(r.URL.Query().Get("ids"), ",")
		result := make(map[string]any)
		for _, id := range ids {
			if id != "" {
				result[id] = map[string]any{"usd": 1000, "usd_24h_change": 0.5}
			}
		}
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(result)
	})
	t.Cleanup(srv.Close)
	return srv
}

func TestWatch_DemoPlanRejected(t *testing.T) {
	srv := simplePriceServer(t)
	withTestClient(t, srv, "demo")

	withFakeStreamer(t, &fakeStreamer{
		connectFn: func(ctx context.Context) (<-chan *ws.CoinUpdate, error) {
			return nil, model.ErrPlanRestricted
		},
	})

	_, _, err := executeCommand(t, "watch", "--ids", "bitcoin", "-o", "json")
	require.Error(t, err)
	assert.ErrorIs(t, err, model.ErrPlanRestricted)
}

func TestWatch_DryRun(t *testing.T) {
	origLoad := loadConfig
	loadConfig = func() (*config.Config, error) {
		return &config.Config{APIKey: "CG-abcdef1234567890", Tier: "paid"}, nil
	}
	t.Cleanup(func() { loadConfig = origLoad })

	stdout, _, err := executeCommand(t, "watch", "--ids", "bitcoin,ethereum", "--dry-run", "-o", "json")
	require.NoError(t, err)

	var out dryRunWSOutput
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	assert.Equal(t, "websocket", out.Transport)
	assert.Contains(t, out.URL, "stream.coingecko.com")
	assert.Contains(t, out.URL, "x_cg_pro_api_key=")
	// Key should be masked.
	assert.NotContains(t, out.URL, "CG-abcdef1234567890")
}

func TestWatch_JSONOutput(t *testing.T) {
	srv := simplePriceServer(t)
	withTestClient(t, srv, "paid")

	withFakeStreamer(t, &fakeStreamer{
		updates: []*ws.CoinUpdate{
			{CoinID: "bitcoin", Price: 45000, Change24h: 2.5, MarketCap: 850e9, Volume24h: 30e9, UpdatedAt: 1700000000},
			{CoinID: "ethereum", Price: 3200, Change24h: -1.2, MarketCap: 380e9, Volume24h: 15e9, UpdatedAt: 1700000000},
		},
	})

	stdout, _, err := executeCommand(t, "watch", "--ids", "bitcoin,ethereum", "-o", "json")
	require.NoError(t, err)

	// Output should be NDJSON — two lines.
	lines := splitNonEmpty(stdout)
	require.Len(t, lines, 2)

	var u1, u2 ws.CoinUpdate
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &u1))
	require.NoError(t, json.Unmarshal([]byte(lines[1]), &u2))
	assert.Equal(t, "bitcoin", u1.CoinID)
	assert.Equal(t, "ethereum", u2.CoinID)
}

func TestWatch_AllIDsInvalid(t *testing.T) {
	// Server returns empty response for unknown IDs.
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{}`))
	})
	defer srv.Close()
	withTestClient(t, srv, "paid")

	_, _, err := executeCommand(t, "watch", "--ids", "notacoin,fakecoin", "-o", "json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "none of the provided coins could be found")
}

func TestWatch_PartialIDsInvalid(t *testing.T) {
	// Server only recognizes "bitcoin", not "fakecoin".
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"bitcoin":{"usd":45000,"usd_24h_change":2.5}}`))
	})
	defer srv.Close()
	withTestClient(t, srv, "paid")

	withFakeStreamer(t, &fakeStreamer{
		updates: []*ws.CoinUpdate{
			{CoinID: "bitcoin", Price: 45000},
		},
	})

	stdout, stderr, err := executeCommand(t, "watch", "--ids", "bitcoin,fakecoin", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, stderr, "not found")
	assert.Contains(t, stderr, "fakecoin")

	lines := splitNonEmpty(stdout)
	require.Len(t, lines, 1)
}

func TestWatch_DryRunShowsPreflights(t *testing.T) {
	origLoad := loadConfig
	loadConfig = func() (*config.Config, error) {
		return &config.Config{APIKey: "test-key", Tier: "paid"}, nil
	}
	t.Cleanup(func() { loadConfig = origLoad })

	stdout, _, err := executeCommand(t, "watch", "--ids", "bitcoin", "--symbols", "eth", "--dry-run", "-o", "json")
	require.NoError(t, err)

	var out dryRunWSOutput
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))
	// Should have two preflight requests: one for --ids validation, one for --symbols resolution.
	assert.Len(t, out.PreflightRequests, 2)
}

func splitNonEmpty(s string) []string {
	var result []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}
