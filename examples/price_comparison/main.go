package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mdnmdn/bits/pkg/bits"
	"github.com/mdnmdn/bits/pkg/config"
)

func main() {
	ctx := context.Background()

	// Initialize the library with a manual configuration.
	// We enable spot markets for Binance and Bitget.
	cfg := &config.Config{
		Binance: config.BinanceConfig{
			Spot: config.MarketConfig{Enabled: true},
		},
		Bitget: config.BitgetConfig{
			Spot: config.MarketConfig{Enabled: true},
		},
		WhiteBit: config.WhiteBitConfig{
			Spot: config.MarketConfig{Enabled: true},
		},
	}

	client := bits.NewClient(cfg)

	// Compare BTC prices across multiple exchanges concurrently.
	symbol := "BTCUSDT"
	exchanges := []string{"binance", "bitget", "whitebit"}

	fmt.Printf("Comparing prices for %s...\n\n", symbol)
	results, err := client.ComparePrices(ctx, symbol, exchanges)
	if err != nil {
		log.Fatalf("Critical error during comparison: %v", err)
	}

	for _, res := range results {
		if len(res.Errors) > 0 {
			fmt.Printf("- %-10s: Error: %v\n", res.Provider, res.Errors[0].Err)
			continue
		}
		fmt.Printf("- %-10s: %.2f (Market: %s)\n", res.Provider, res.Data.Price, res.Market)
	}
}
