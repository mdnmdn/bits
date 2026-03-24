package provider

import (
	"fmt"

	"github.com/coingecko/coingecko-cli/internal/config"
	"github.com/coingecko/coingecko-cli/internal/provider/binance"
	"github.com/coingecko/coingecko-cli/internal/provider/bitget"
	"github.com/coingecko/coingecko-cli/internal/provider/coingecko"
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
