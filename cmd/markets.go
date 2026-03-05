package cmd

import (
	"context"
	"fmt"
	"strings"

	"coingecko-cli/internal/api"
	"coingecko-cli/internal/config"
	"coingecko-cli/internal/display"
	"coingecko-cli/internal/export"

	"github.com/spf13/cobra"
)

var marketsCmd = &cobra.Command{
	Use:   "markets",
	Short: "List top coins by market cap",
	RunE:  runMarkets,
}

func init() {
	marketsCmd.Flags().Int("total", 100, "Total number of coins to fetch")
	marketsCmd.Flags().String("vs", "usd", "Target currency")
	marketsCmd.Flags().String("order", "market_cap_desc", "Sort order")
	marketsCmd.Flags().String("category", "", "Filter by category")
	marketsCmd.Flags().String("export", "", "Export to CSV file path")
	rootCmd.AddCommand(marketsCmd)
}

func runMarkets(cmd *cobra.Command, args []string) error {
	total, _ := cmd.Flags().GetInt("total")
	vs, _ := cmd.Flags().GetString("vs")
	order, _ := cmd.Flags().GetString("order")
	category, _ := cmd.Flags().GetString("category")
	exportPath, _ := cmd.Flags().GetString("export")
	jsonOut := outputJSON(cmd)

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	client := api.NewClient(cfg)
	ctx := context.Background()

	var allCoins []api.MarketCoin
	perPage := 250
	remaining := total

	for page := 1; remaining > 0; page++ {
		fetch := perPage
		if remaining < perPage {
			fetch = remaining
		}
		coins, err := client.CoinMarkets(ctx, vs, fetch, page, order, category)
		if err != nil {
			return err
		}
		allCoins = append(allCoins, coins...)
		remaining -= len(coins)
		if len(coins) < fetch {
			break
		}
	}

	if jsonOut {
		printJSONRaw(allCoins)
		return nil
	}

	headers := []string{"Rank", "Name", "Symbol", "Price", "Market Cap", "Volume", "24h Change"}
	rows := make([][]string, len(allCoins))
	for i, c := range allCoins {
		rows[i] = []string{
			fmt.Sprintf("%d", c.MarketCapRank),
			c.Name,
			strings.ToUpper(c.Symbol),
			display.FormatPrice(c.CurrentPrice),
			display.FormatLargeNumber(c.MarketCap),
			display.FormatLargeNumber(c.TotalVolume),
			display.ColorPercent(c.PriceChangePercentage24h),
		}
	}

	display.PrintTable(headers, rows)

	if exportPath != "" {
		plainRows := make([][]string, len(allCoins))
		for i, c := range allCoins {
			plainRows[i] = []string{
				fmt.Sprintf("%d", c.MarketCapRank),
				c.Name,
				strings.ToUpper(c.Symbol),
				fmt.Sprintf("%.8f", c.CurrentPrice),
				fmt.Sprintf("%.2f", c.MarketCap),
				fmt.Sprintf("%.2f", c.TotalVolume),
				fmt.Sprintf("%.2f", c.PriceChangePercentage24h),
			}
		}
		if err := export.ExportCSV(exportPath, headers, plainRows); err != nil {
			return err
		}
		warnf("Exported to %s\n", exportPath)
	}

	return nil
}
