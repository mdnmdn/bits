package provider

import (
	"fmt"

	"github.com/mdnmdn/bits/internal/config"
	"github.com/mdnmdn/bits/internal/provider/binance"
	"github.com/mdnmdn/bits/internal/provider/bitget"
	"github.com/mdnmdn/bits/internal/provider/coingecko"
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
