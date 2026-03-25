package cmd

import (
	"context"

	"github.com/mdnmdn/bits/internal/config"
	"github.com/mdnmdn/bits/internal/model"
	"github.com/mdnmdn/bits/internal/provider"
	"github.com/mdnmdn/bits/internal/ws"
	"github.com/spf13/cobra"
)

// userAgent is the User-Agent header sent with all API and WebSocket requests.
var userAgent = "coingecko-cli/" + version

// providerOverride is set by the --provider flag via PersistentPreRun on rootCmd.
// Empty means use config default.
var providerOverride string

// marketTypeOverride is set by the --market-type flag via PersistentPreRun on rootCmd.
var marketTypeOverride string

func init() {
	// Resolve --provider flag before any subcommand runs.
	existing := rootCmd.PersistentPreRun
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if p, _ := cmd.Flags().GetString("provider"); p != "" {
			providerOverride = p
		}
		if m, _ := cmd.Flags().GetString("market-type"); m != "" {
			marketTypeOverride = m
		}
		if existing != nil {
			existing(cmd, args)
		}
	}
}

// newAPIClient is the factory used by command handlers to create API clients.
// Tests override this to inject httptest-backed clients.
var newAPIClient = func(cfg *config.Config) provider.Provider {
	name := cfg.ActiveProvider()
	if providerOverride != "" {
		name = providerOverride
	}

	// Apply market type override to the relevant config section
	if marketTypeOverride != "" {
		cfg.Binance.MarketType = marketTypeOverride
		cfg.Bitget.MarketType = marketTypeOverride
	}

	c, err := provider.NewProvider(name, cfg)
	if err != nil {
		// Fall back to coingecko on unknown provider.
		c, _ = provider.NewProvider("coingecko", cfg)
	}
	c.SetUserAgent(userAgent)
	return c
}

// loadConfig is the function used by command handlers to load configuration.
// Tests override this to inject test configs without touching the real config file.
var loadConfig = config.Load

// Streamer abstracts the WebSocket streaming client for testability.
type Streamer interface {
	Connect(ctx context.Context) (<-chan *model.CoinUpdate, error)
	Close() error
}

// newStreamer is the factory used by command handlers to create WebSocket clients.
// Tests override this to inject test doubles.
var newStreamer = func(cfg *config.Config, coinIDs []string) Streamer {
	c := ws.NewClient(cfg, coinIDs)
	c.UserAgent = userAgent
	return c
}
