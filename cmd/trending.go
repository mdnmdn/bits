package cmd

import (
	"fmt"

	"github.com/coingecko/coingecko-cli/internal/api"
	"github.com/coingecko/coingecko-cli/internal/config"
	"github.com/coingecko/coingecko-cli/internal/display"

	"github.com/spf13/cobra"
)

var trendingCmd = &cobra.Command{
	Use:   "trending",
	Short: "Show trending coins, NFTs, and categories",
	Example: `  cg trending
  cg trending --show-max coins,nfts,categories
  cg trending -o json`,
	RunE: runTrending,
}

func init() {
	trendingCmd.Flags().String("show-max", "", "Show max results for types: coins, nfts, categories (comma-separated, paid plans only)")
	rootCmd.AddCommand(trendingCmd)
}

func runTrending(cmd *cobra.Command, args []string) error {
	showMax, _ := cmd.Flags().GetString("show-max")
	jsonOut := outputJSON(cmd)

	if !jsonOut {
		display.PrintBanner()
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if showMax != "" && !cfg.IsPaid() {
		return fmt.Errorf("--show-max: %w", api.ErrPlanRestricted)
	}

	if isDryRun(cmd) {
		params := map[string]string{}
		if showMax != "" {
			params["show_max"] = showMax
		}
		return printDryRun(cfg, "trending", "/search/trending", params, nil)
	}

	client := api.NewClient(cfg)
	ctx := cmd.Context()

	resp, err := client.SearchTrending(ctx, showMax)
	if err != nil {
		return err
	}

	if jsonOut {
		return printJSONRaw(resp)
	}

	// Trending Coins
	fmt.Println("Trending Coins")
	fmt.Println()
	coinHeaders := []string{"#", "Name", "Symbol", "Market Cap Rank"}
	coinRows := make([][]string, 0, len(resp.Coins))
	for i, c := range resp.Coins {
		coinRows = append(coinRows, []string{
			fmt.Sprintf("%d", i+1),
			display.SanitizeCell(c.Item.Name),
			display.SanitizeCell(c.Item.Symbol),
			display.FormatRank(c.Item.MarketCapRank),
		})
	}
	display.PrintTable(coinHeaders, coinRows)

	// Trending NFTs
	if len(resp.NFTs) > 0 {
		fmt.Println()
		fmt.Println("Trending NFTs")
		fmt.Println()
		nftHeaders := []string{"#", "Name", "Symbol", "Floor Price 24h Change"}
		nftRows := make([][]string, 0, len(resp.NFTs))
		for i, n := range resp.NFTs {
			nftRows = append(nftRows, []string{
				fmt.Sprintf("%d", i+1),
				display.SanitizeCell(n.Name),
				display.SanitizeCell(n.Symbol),
				display.ColorPercent(n.FloorPriceInUSD24hPC),
			})
		}
		display.PrintTable(nftHeaders, nftRows)
	}

	// Trending Categories
	if len(resp.Categories) > 0 {
		fmt.Println()
		fmt.Println("Trending Categories")
		fmt.Println()
		catHeaders := []string{"#", "Name", "Market Cap 1h Change"}
		catRows := make([][]string, 0, len(resp.Categories))
		for i, cat := range resp.Categories {
			catRows = append(catRows, []string{
				fmt.Sprintf("%d", i+1),
				display.SanitizeCell(cat.Name),
				display.ColorPercent(cat.MarketCap1hChange),
			})
		}
		display.PrintTable(catHeaders, catRows)
	}

	return nil
}
