package provider

import (
	"fmt"

	"github.com/coingecko/coingecko-cli/internal/config"
	"github.com/coingecko/coingecko-cli/internal/provider/coingecko"
)

func NewProvider(name string, cfg *config.Config) (Provider, error) {
	switch name {
	case "coingecko":
		return coingecko.NewClient(cfg), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
}
