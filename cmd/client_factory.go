package cmd

import (
	"context"

	"github.com/coingecko/coingecko-cli/internal/config"
	"github.com/coingecko/coingecko-cli/internal/provider"
	"github.com/coingecko/coingecko-cli/internal/ws"
)

// userAgent is the User-Agent header sent with all API and WebSocket requests.
var userAgent = "coingecko-cli/" + version

// newAPIClient is the factory used by command handlers to create API clients.
// Tests override this to inject httptest-backed clients.
var newAPIClient = func(cfg *config.Config) provider.Provider {
	// Default to coingecko provider for now.
	c, _ := provider.NewProvider("coingecko", cfg)
	c.SetUserAgent(userAgent)
	return c
}


// loadConfig is the function used by command handlers to load configuration.
// Tests override this to inject test configs without touching the real config file.
var loadConfig = config.Load

// Streamer abstracts the WebSocket streaming client for testability.
type Streamer interface {
	Connect(ctx context.Context) (<-chan *ws.CoinUpdate, error)
	Close() error
}

// newStreamer is the factory used by command handlers to create WebSocket clients.
// Tests override this to inject test doubles.
var newStreamer = func(cfg *config.Config, coinIDs []string) Streamer {
	c := ws.NewClient(cfg, coinIDs)
	c.UserAgent = userAgent
	return c
}
