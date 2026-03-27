package cmd

import (
	"fmt"

	"github.com/mdnmdn/bits/internal/display"
	"github.com/mdnmdn/bits/internal/provider"

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

	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	client, err := newAPIClientWithFeature("markets")(cfg)
	if err != nil {
		return err
	}
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

	ml, ok := client.(provider.MarketLister)
	if !ok {
		return fmt.Errorf("%s provider does not support market listings", client.ID())
	}
	allCoins, err := ml.FetchAllMarkets(ctx, vs, total, order, category)
	if err != nil {
		return err
	}

	if jsonOut {
		return printJSONRaw(allCoins)
	}

	headers := []string{"Rank", "Name", "Symbol", "Price", "Market Cap", "Volume", "24h Change"}
	rows := make([][]string, len(allCoins))
	var csvRows [][]string
	if exportPath != "" {
		csvRows = make([][]string, len(allCoins))
	}
	for i, c := range allCoins {
		rank := fmt.Sprintf("%d", c.MarketCapRank)
		name := display.SanitizeCell(c.Name)
		symbol := display.FormatSymbol(c.Symbol)
		rows[i] = []string{
			rank, name, symbol,
			display.FormatPrice(c.CurrentPrice, vs),
			display.FormatLargeNumber(c.MarketCap, vs),
			display.FormatLargeNumber(c.TotalVolume, vs),
			display.ColorPercent(c.PriceChangePercentage24h),
		}
		if exportPath != "" {
			csvRows[i] = []string{
				rank, name, symbol,
				fmt.Sprintf("%.8f", c.CurrentPrice),
				fmt.Sprintf("%.2f", c.MarketCap),
				fmt.Sprintf("%.2f", c.TotalVolume),
				fmt.Sprintf("%.2f", c.PriceChangePercentage24h),
			}
		}
	}

	display.PrintTable(headers, rows)

	if exportPath != "" {
		if err := exportCSV(exportPath, headers, csvRows); err != nil {
			return err
		}
	}

	return nil
}
