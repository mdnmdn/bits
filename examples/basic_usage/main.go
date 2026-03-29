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

	// Initialize the library with default configuration.
	// You can also load it from a file or environment.
	cfg := &config.Config{
		Binance: config.BinanceConfig{
			Spot: config.MarketConfig{Enabled: true},
		},
	}

	client := bits.NewClient(cfg)

	// Fetch the current price of Ethereum on Binance.
	symbol := "ETHUSDT"
	providerID := "binance"

	fmt.Printf("Fetching price for %s from %s...\n", symbol, providerID)
	res, err := client.GetPrice(ctx, symbol, providerID)
	if err != nil {
		log.Fatalf("Failed to get price: %v", err)
	}

	fmt.Printf("Price: %.2f (Provider: %s, Market: %s)\n", res.Data.Price, res.Provider, res.Market)
}
