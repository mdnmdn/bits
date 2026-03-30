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
	}

	fmt.Println("=== WITHOUT WithSymbolEngine() ===")

	// WITHOUT WithSymbolEngine() - still works
	client := bits.NewClient(cfg)

	// ResolveSymbol still works (creates engine on demand)
	sym, err := client.ResolveSymbol(ctx, "ETH-USDT", "binance", model.MarketSpot)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ResolveSymbol: %s\n", sym)

	// NormalizeSymbol works without any engine
	fmt.Printf("NormalizeSymbol: %s\n", bits.NormalizeSymbol("BTC_USDT"))

	// GetPrice works (just passes through)
	price, err := client.GetPrice(ctx, "ETHUSDT", "binance")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("GetPrice (manual symbol): $%.2f\n", price.Data.Price)

	fmt.Println("\n=== WITH WithSymbolEngine() ===")

	// WITH WithSymbolEngine() - shared cache
	client2 := bits.NewClient(cfg, bits.WithSymbolEngine())

	sym2, _ := client2.ResolveSymbol(ctx, "ETH-USDT", "binance", model.MarketSpot)
	fmt.Printf("ResolveSymbol: %s\n", sym2)

	price2, _ := client2.GetPriceWithResolution(ctx, "ETH-USDT", "binance", model.MarketSpot)
	fmt.Printf("GetPriceWithResolution: $%.2f\n", price2.Data.Price)
}
