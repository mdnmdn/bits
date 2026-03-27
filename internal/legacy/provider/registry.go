package provider

import (
	"fmt"

	"github.com/mdnmdn/bits/internal/capability"
	"github.com/mdnmdn/bits/internal/config"
	"github.com/mdnmdn/bits/internal/legacy/provider/binance"
	"github.com/mdnmdn/bits/internal/legacy/provider/bitget"
	"github.com/mdnmdn/bits/internal/legacy/provider/coingecko"
)

// AvailableProviders lists all supported provider names.
var AvailableProviders = []string{"coingecko", "binance", "bitget"}

func NewProvider(name string, cfg *config.Config) (Provider, error) {
	switch name {
	case "coingecko":
		return coingecko.NewClient(cfg), nil
	case "binance":
		return binance.NewClient(cfg.Binance), nil
	case "bitget":
		return bitget.NewClient(cfg.Bitget), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s (available: coingecko, binance, bitget)", name)
	}
}

// AllCapabilities returns the CapabilityMatrix for every registered provider.
// Providers are instantiated with a zero-value config because Capabilities() is
// a pure static declaration that makes no network calls.
func AllCapabilities() map[string]capability.CapabilityMatrix {
	result := make(map[string]capability.CapabilityMatrix, len(AvailableProviders))
	emptyCfg := &config.Config{}
	for _, name := range AvailableProviders {
		p, err := NewProvider(name, emptyCfg)
		if err != nil {
			continue
		}
		if cp, ok := p.(capability.CapabilityProvider); ok {
			result[name] = cp.Capabilities()
		}
	}
	return result
}
