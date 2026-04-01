package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mdnmdn/bits"
	"github.com/mdnmdn/bits/config"
	"github.com/mdnmdn/bits/model"
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

	{
		fmt.Println("=== Multi-Provider Client ===")
		client := bits.NewClient(cfg, bits.WithSymbolEngine())

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

	{
		fmt.Println("\n=== Provider-Specific Clients ===")
		binance := bits.NewProvider(cfg, "binance")
		bitget := bits.NewProvider(cfg, "bitget")
		whitebit := bits.NewProvider(cfg, "whitebit")

		exchanges := []*bits.Client{binance, bitget, whitebit}
		symbol := "BTCUSDT"

		for _, p := range exchanges {
			fmt.Printf("\n%s:\n", p.ID())
			fmt.Printf("  Capabilities: %v\n", p.Capabilities())

			res, err := p.Price(ctx, []string{symbol}, "")
			if err != nil {
				fmt.Printf("  Price: Error: %v\n", err)
				continue
			}
			if len(res.Data) == 0 {
				fmt.Printf("  Price: Error: %v\n", res.Errors[0].Err)
				continue
			}
			fmt.Printf("  Price: $%.2f\n", res.Data[0].Price)
		}
	}
}
