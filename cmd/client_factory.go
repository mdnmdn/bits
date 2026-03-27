package cmd

import (
	"context"
	"fmt"

	"github.com/mdnmdn/bits/internal/capability"
	"github.com/mdnmdn/bits/internal/config"
	"github.com/mdnmdn/bits/internal/model"
	"github.com/mdnmdn/bits/internal/provider"
	"github.com/mdnmdn/bits/internal/provider/binance"
	"github.com/mdnmdn/bits/internal/provider/bitget"
	"github.com/mdnmdn/bits/internal/ws"
	"github.com/spf13/cobra"
)

var userAgent = "coingecko-cli/" + version

var providerOverride string
var marketTypeOverride string

var commandFeatureMap = map[string]capability.Feature{
	"price":              capability.FeaturePrice,
	"markets":            capability.FeatureMarketsList,
	"search":             capability.FeatureSearch,
	"trending":           capability.FeatureTrending,
	"history":            capability.FeatureHistorical,
	"top-gainers-losers": capability.FeatureGainersLosers,
	"ticker":             capability.FeatureTicker24h,
	"orderbook":          capability.FeatureOrderBook,
	"watch":              capability.FeatureStreamPrice,
	"tui-markets":        capability.FeatureMarketsList,
	"tui-trending":       capability.FeatureTrending,
}

func init() {
	existing := RootCmd.PersistentPreRun
	RootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
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

func resolveProvider(cfg *config.Config, requiredFeature capability.Feature) (provider.Provider, error) {
	providerName := cfg.ActiveProvider()
	if providerOverride != "" {
		providerName = providerOverride
	}

	allCaps := provider.AllCapabilities()

	if requiredFeature != "" {
		if providerOverride != "" {
			if caps, ok := allCaps[providerName]; ok {
				key := capability.CapabilityKey{Market: capability.MarketSpot, Feature: requiredFeature}
				if caps[key] {
					return createProviderClient(providerName, cfg)
				}
			}
			return nil, fmt.Errorf("provider %q does not support %s", providerName, requiredFeature)
		}

		if caps, ok := allCaps[providerName]; ok {
			key := capability.CapabilityKey{Market: capability.MarketSpot, Feature: requiredFeature}
			if caps[key] {
				return createProviderClient(providerName, cfg)
			}
		}

		for _, name := range provider.AvailableProviders {
			if caps, ok := allCaps[name]; ok {
				key := capability.CapabilityKey{Market: capability.MarketSpot, Feature: requiredFeature}
				if caps[key] {
					return createProviderClient(name, cfg)
				}
			}
		}
		return nil, fmt.Errorf("no provider supports %s", requiredFeature)
	}

	return createProviderClient(providerName, cfg)
}

func createProviderClient(name string, cfg *config.Config) (provider.Provider, error) {
	c, err := provider.NewProvider(name, cfg)
	if err != nil {
		c, _ = provider.NewProvider("coingecko", cfg)
		return c, nil
	}

	if marketTypeOverride != "" {
		if bc, ok := c.(*binance.Client); ok {
			bc.SetMarketType(marketTypeOverride)
		} else if bgc, ok := c.(*bitget.Client); ok {
			bgc.SetMarketType(marketTypeOverride)
		}
	}

	c.SetUserAgent(userAgent)
	return c, nil
}

var newAPIClient = func(cfg *config.Config) provider.Provider {
	c, _ := resolveProvider(cfg, "")
	return c
}

var newAPIClientWithFeature = func(cmdName string) func(*config.Config) (provider.Provider, error) {
	return func(cfg *config.Config) (provider.Provider, error) {
		feature := commandFeatureMap[cmdName]
		return resolveProvider(cfg, feature)
	}
}

var loadedConfigPath string

var loadConfig = func() (*config.Config, error) {
	cfg, path, err := config.Load()
	loadedConfigPath = path
	return cfg, err
}

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
