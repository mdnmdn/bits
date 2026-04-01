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
		WhiteBit: config.WhiteBitConfig{
			Spot: config.MarketConfig{Enabled: true},
		},
	}

	{
		fmt.Println("=== Multi-Provider Client ===")
		client := bits.NewClient(cfg, bits.WithSymbolEngine())

		normalizedInput := "ETH-USDT"
		binanceSymbol, err := client.ResolveSymbol(ctx, normalizedInput, "binance", model.MarketSpot)
		if err != nil {
			log.Fatalf("Failed to resolve for binance: %v", err)
		}
		fmt.Printf("ResolveSymbol('%s', 'binance', spot): %s\n", normalizedInput, binanceSymbol)

		price, err := client.GetPriceWithResolution(ctx, "ETH-USDT", "binance", model.MarketSpot)
		if err != nil {
			log.Fatalf("Failed to get price: %v", err)
		}
		fmt.Printf("ETH-USDT on binance: $%.2f\n", price.Data.Price)

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

	{
		fmt.Println("\n=== Provider-Specific Client ===")
		binance := bits.NewProvider(cfg, "binance")
		fmt.Printf("Provider ID: %s\n", binance.ID())
		fmt.Printf("Capabilities: %v\n", binance.Capabilities())

		res, err := binance.Price(ctx, []string{"BTCUSDT"}, "")
		if err != nil {
			log.Fatalf("Failed to get price: %v", err)
		}
		fmt.Printf("BTCUSDT price: $%.2f\n", res.Data[0].Price)

		ticker, err := binance.Ticker24h(ctx, "BTCUSDT", model.MarketSpot)
		if err != nil {
			log.Fatalf("Failed to get ticker: %v", err)
		}
		fmt.Printf("BTCUSDT 24h: high=$%.2f low=$%.2f vol=%.2f\n",
			*ticker.Data.HighPrice, *ticker.Data.LowPrice, *ticker.Data.Volume)

		candles, err := binance.Candles(ctx, "BTCUSDT", model.MarketSpot, "1h", model.CandleOpts{})
		if err != nil {
			log.Fatalf("Failed to get candles: %v", err)
		}
		if len(candles.Data) > 0 {
			fmt.Printf("BTCUSDT last candle: open=$%.2f close=$%.2f\n",
				candles.Data[0].Open, candles.Data[len(candles.Data)-1].Close)
		}
	}

	{
		fmt.Println("\n=== Symbol Normalization ===")
		fmt.Printf("NormalizeSymbol('BTCUSDT'):  %s\n", bits.NormalizeSymbol("BTCUSDT"))
		fmt.Printf("NormalizeSymbol('BTC_USDT'): %s\n", bits.NormalizeSymbol("BTC_USDT"))
		fmt.Printf("NormalizeSymbol('btc-usdt'): %s\n", bits.NormalizeSymbol("btc-usdt"))
	}
}
