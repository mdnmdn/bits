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
	}

	{
		fmt.Println("=== WITHOUT WithSymbolEngine() ===")
		client := bits.NewClient(cfg)

		sym, err := client.ResolveSymbol(ctx, "ETH-USDT", "binance", model.MarketSpot)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ResolveSymbol: %s\n", sym)

		fmt.Printf("NormalizeSymbol: %s\n", bits.NormalizeSymbol("BTC_USDT"))

		price, err := client.GetPrice(ctx, "ETHUSDT", "binance")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("GetPrice (manual symbol): $%.2f\n", price.Data.Price)
	}

	{
		fmt.Println("\n=== WITH WithSymbolEngine() ===")
		client := bits.NewClient(cfg, bits.WithSymbolEngine())

		sym, _ := client.ResolveSymbol(ctx, "ETH-USDT", "binance", model.MarketSpot)
		fmt.Printf("ResolveSymbol: %s\n", sym)

		price, _ := client.GetPriceWithResolution(ctx, "ETH-USDT", "binance", model.MarketSpot)
		fmt.Printf("GetPriceWithResolution: $%.2f\n", price.Data.Price)
	}

	{
		fmt.Println("\n=== Provider-Specific Client ===")
		p := bits.NewProvider(cfg, "binance")

		sym, _ := p.ResolveSymbol(ctx, "ETH-USDT", "binance", model.MarketSpot)
		fmt.Printf("ResolveSymbol: %s\n", sym)

		price, _ := p.Price(ctx, []string{"ETHUSDT"}, "")
		fmt.Printf("Price: $%.2f\n", price.Data[0].Price)
	}
}
