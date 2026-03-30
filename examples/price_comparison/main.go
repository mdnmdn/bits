package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mdnmdn/bits/pkg/bits"
	"github.com/mdnmdn/bits/pkg/config"
	"github.com/mdnmdn/bits/pkg/model"
)

func main() {
	ctx := context.Background()

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

	// Use WithSymbolEngine() to enable automatic symbol resolution
	// This handles different symbol formats automatically
	client := bits.NewClient(cfg, bits.WithSymbolEngine())

	// Compare BTC prices across multiple exchanges
	// With the engine, we can use normalized symbol format for all exchanges
	normalizedSymbol := "BTC-USDT"
	exchanges := []string{"binance", "bitget", "whitebit"}

	fmt.Printf("Comparing prices for %s across exchanges...\n\n", normalizedSymbol)
	results, err := client.ComparePricesWithResolution(ctx, normalizedSymbol, exchanges, model.MarketSpot)
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
