package cmd

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/coingecko/coingecko-cli/internal/api"
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

func TestWatch_DemoPlanRejected(t *testing.T) {
	origLoad := loadConfig
	loadConfig = func() (*config.Config, error) {
		return &config.Config{APIKey: "test-key", Tier: "demo"}, nil
	}
	t.Cleanup(func() { loadConfig = origLoad })

	withFakeStreamer(t, &fakeStreamer{
		connectFn: func(ctx context.Context) (<-chan *ws.CoinUpdate, error) {
			return nil, api.ErrPlanRestricted
		},
	})

	_, _, err := executeCommand(t, "watch", "--ids", "bitcoin", "-o", "json")
	require.Error(t, err)
	assert.ErrorIs(t, err, api.ErrPlanRestricted)
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
	origLoad := loadConfig
	loadConfig = func() (*config.Config, error) {
		return &config.Config{APIKey: "test-key", Tier: "paid"}, nil
	}
	t.Cleanup(func() { loadConfig = origLoad })

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
