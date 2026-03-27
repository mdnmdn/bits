// Package registry wires the concrete provider implementations into the
// provider.Provider interface. It lives in its own package to avoid the
// import cycle that would arise if internal/provider imported its own
// sub-packages.
package registry

import (
	"fmt"

	"github.com/mdnmdn/bits/internal/config"
	"github.com/mdnmdn/bits/internal/provider"
	"github.com/mdnmdn/bits/internal/provider/binance"
	"github.com/mdnmdn/bits/internal/provider/bitget"
	"github.com/mdnmdn/bits/internal/provider/coingecko"
)

// NewProvider constructs a provider by name using the given config.
func NewProvider(name string, cfg *config.Config) (provider.Provider, error) {
	switch name {
	case "coingecko", "":
		return coingecko.NewClient(cfg), nil
	case "binance":
		return binance.NewClient(cfg.Binance), nil
	case "bitget":
		return bitget.NewClient(cfg.Bitget), nil
	default:
		return nil, fmt.Errorf("unknown provider %q", name)
	}
}

// AllProviderIDs returns the IDs of all registered providers.
func AllProviderIDs() []string {
	return []string{"coingecko", "binance", "bitget"}
}
