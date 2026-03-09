package cmd

import (
	"github.com/coingecko/coingecko-cli/internal/api"
	"github.com/coingecko/coingecko-cli/internal/config"
)

// newAPIClient is the factory used by command handlers to create API clients.
// Tests override this to inject httptest-backed clients.
var newAPIClient = func(cfg *config.Config) *api.Client {
	return api.NewClient(cfg)
}

// loadConfig is the function used by command handlers to load configuration.
// Tests override this to inject test configs without touching the real config file.
var loadConfig = config.Load
