// Package registry wires the concrete provider implementations into the
// provider.Provider interface. It lives in its own package to avoid the
// import cycle that would arise if internal/provider imported its own
// sub-packages.
package registry

import (
	"fmt"
	"strings"

	"github.com/mdnmdn/bits/config"
	"github.com/mdnmdn/bits/provider"
	"github.com/mdnmdn/bits/provider/binance"
	"github.com/mdnmdn/bits/provider/bitget"
	"github.com/mdnmdn/bits/provider/coingecko"
	"github.com/mdnmdn/bits/provider/cryptocom"
	"github.com/mdnmdn/bits/provider/mexc"
	"github.com/mdnmdn/bits/provider/whitebit"
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
	// cryptocom aliases
	"cdc": "cryptocom",
	"cro": "cryptocom",
	"crypto.com": "cryptocom",
	"crypto": "cryptocom",
	// mexc aliases
	"mx": "mexc",
	"mxc": "mexc",
}

// ResolveProvider resolves an alias to the canonical provider name.
// If the name is already canonical or unknown, it is returned as-is.
// Case-insensitive.
func ResolveProvider(name string) string {
	lower := strings.ToLower(name)
	if canonical, ok := providerAliases[lower]; ok {
		return canonical
	}
	return lower
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
	case "cryptocom":
		return cryptocom.NewClient(cfg.CryptoCom), nil
	case "mexc":
		return mexc.NewClient(cfg.MEXC), nil
	default:
		return nil, fmt.Errorf("unknown provider %q", name)
	}
}

// AllProviderIDs returns the IDs of all registered providers.
func AllProviderIDs() []string {
	return []string{"coingecko", "binance", "bitget", "whitebit", "cryptocom", "mexc"}
}

// AllProviderIDsWithAliases returns all provider IDs including aliases.
func AllProviderIDsWithAliases() []string {
	return []string{
		"coingecko", "cg",
		"binance", "bn",
		"bitget", "bg",
		"whitebit", "wb",
		"cryptocom", "cdc", "cro",
		"mexc", "mx",
	}
}
