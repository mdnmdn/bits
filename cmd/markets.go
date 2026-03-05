package cmd

import (
	"fmt"
	"strings"

	"github.com/coingecko/coingecko-cli/internal/api"
	"github.com/coingecko/coingecko-cli/internal/config"
	"github.com/coingecko/coingecko-cli/internal/display"

	"github.com/spf13/cobra"
)

var marketsCmd = &cobra.Command{
	Use:   "markets",
	Short: "List top coins by market cap",
	Example: `  cg markets
  cg markets --total 20 --vs eur
  cg markets --category layer-1 --export coins.csv`,
	RunE: runMarkets,
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

	if !jsonOut {
		display.PrintBanner()
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	client := api.NewClient(cfg)
	ctx := cmd.Context()

	perPage := 250

	if isDryRun(cmd) {
		params := map[string]string{
			"vs_currency": vs,
			"per_page":    fmt.Sprintf("%d", perPage),
			"page":        "1",
			"order":       order,
		}
		if category != "" {
			params["category"] = category
		}
		pages := (total + perPage - 1) / perPage
		return printDryRun(cfg, "markets", "/coins/markets", params, &paginationInfo{
			TotalRequested: total,
			PerPage:        perPage,
			Pages:          pages,
		})
	}

	var allCoins []api.MarketCoin
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
		return printJSONRaw(allCoins)
	}

	headers := []string{"Rank", "Name", "Symbol", "Price", "Market Cap", "Volume", "24h Change"}
	rows := make([][]string, len(allCoins))
	for i, c := range allCoins {
		rows[i] = []string{
			fmt.Sprintf("%d", c.MarketCapRank),
			display.SanitizeCell(c.Name),
			strings.ToUpper(display.SanitizeCell(c.Symbol)),
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
				display.SanitizeCell(c.Name),
				strings.ToUpper(display.SanitizeCell(c.Symbol)),
				fmt.Sprintf("%.8f", c.CurrentPrice),
				fmt.Sprintf("%.2f", c.MarketCap),
				fmt.Sprintf("%.2f", c.TotalVolume),
				fmt.Sprintf("%.2f", c.PriceChangePercentage24h),
			}
		}
		if err := exportCSV(exportPath, headers, plainRows); err != nil {
			return err
		}
	}

	return nil
}
