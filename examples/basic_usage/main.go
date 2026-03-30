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

	// Initialize the library with default configuration.
	// Use WithSymbolEngine() to enable automatic symbol resolution.
	cfg := &config.Config{
		Binance: config.BinanceConfig{
			Spot: config.MarketConfig{Enabled: true},
		},
		WhiteBit: config.WhiteBitConfig{
			Spot: config.MarketConfig{Enabled: true},
		},
	}

	// Create client with symbol resolution enabled
	client := bits.NewClient(cfg, bits.WithSymbolEngine())

	// Demonstrate symbol normalization
	fmt.Println("=== Symbol Normalization ===")
	fmt.Printf("NormalizeSymbol('BTCUSDT'):  %s\n", bits.NormalizeSymbol("BTCUSDT"))
	fmt.Printf("NormalizeSymbol('BTC_USDT'): %s\n", bits.NormalizeSymbol("BTC_USDT"))
	fmt.Printf("NormalizeSymbol('btc-usdt'): %s\n", bits.NormalizeSymbol("btc-usdt"))

	// Demonstrate symbol resolution
	fmt.Println("\n=== Symbol Resolution ===")
	normalizedInput := "ETH-USDT"

	// Resolve to different provider formats
	binanceSymbol, err := client.ResolveSymbol(ctx, normalizedInput, "binance", model.MarketSpot)
	if err != nil {
		log.Fatalf("Failed to resolve for binance: %v", err)
	}
	fmt.Printf("ResolveSymbol('%s', 'binance', spot): %s\n", normalizedInput, binanceSymbol)

	whitebitSymbol, err := client.ResolveSymbol(ctx, normalizedInput, "whitebit", model.MarketSpot)
	if err != nil {
		log.Fatalf("Failed to resolve for whitebit: %v", err)
	}
	fmt.Printf("ResolveSymbol('%s', 'whitebit', spot): %s\n", normalizedInput, whitebitSymbol)

	// Fetch prices with automatic resolution
	fmt.Println("\n=== Price with Resolution ===")
	price, err := client.GetPriceWithResolution(ctx, "ETH-USDT", "binance", model.MarketSpot)
	if err != nil {
		log.Fatalf("Failed to get price: %v", err)
	}
	fmt.Printf("ETH-USDT on binance: $%.2f\n", price.Data.Price)

	// Compare prices across exchanges with automatic resolution
	fmt.Println("\n=== Price Comparison ===")
	results, err := client.ComparePricesWithResolution(ctx, "BTC-USDT",
		[]string{"binance", "whitebit"}, model.MarketSpot)
	if err != nil {
		log.Fatalf("Failed to compare prices: %v", err)
	}

	for _, res := range results {
		if len(res.Errors) > 0 {
			fmt.Printf("- %-10s: Error: %v\n", res.Provider, res.Errors[0].Err)
			continue
		}
		fmt.Printf("- %-10s: $%.2f\n", res.Provider, res.Data.Price)
	}
}
