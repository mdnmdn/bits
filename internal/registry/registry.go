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
	"github.com/mdnmdn/bits/internal/provider/whitebit"
)

var providerAliases = map[string]string{
	// coingecko aliases
	"cg": "coingecko",
	// binance aliases
	"bn": "binance",
	// bitget aliases
	"bg": "bitget",
	// whitebit aliases
	"wb": "whitebit",
}

// ResolveProvider resolves an alias to the canonical provider name.
// If the name is already canonical or unknown, it is returned as-is.
func ResolveProvider(name string) string {
	if canonical, ok := providerAliases[name]; ok {
		return canonical
	}
	return name
}

// NewProvider constructs a provider by name using the given config.
// Accepts both canonical names and aliases (e.g., "binance" or "bn").
func NewProvider(name string, cfg *config.Config) (provider.Provider, error) {
	name = ResolveProvider(name)

	switch name {
	case "coingecko", "":
		return coingecko.NewClient(cfg), nil
	case "binance":
		return binance.NewClient(cfg.Binance), nil
	case "bitget":
		return bitget.NewClient(cfg.Bitget), nil
	case "whitebit":
		return whitebit.NewClient(cfg.WhiteBit), nil
	default:
		return nil, fmt.Errorf("unknown provider %q", name)
	}
}

// AllProviderIDs returns the IDs of all registered providers.
func AllProviderIDs() []string {
	return []string{"coingecko", "binance", "bitget", "whitebit"}
}

// AllProviderIDsWithAliases returns all provider IDs including aliases.
func AllProviderIDsWithAliases() []string {
	return []string{
		"coingecko", "cg",
		"binance", "bn",
		"bitget", "bg",
		"whitebit", "wb",
	}
}
