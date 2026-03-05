package cmd

import (
	"fmt"

	"coingecko-cli/internal/api"
	"coingecko-cli/internal/config"
	"coingecko-cli/internal/display"

	"github.com/spf13/cobra"
)

var trendingCmd = &cobra.Command{
	Use:   "trending",
	Short: "Show trending coins, NFTs, and categories",
	Example: `  cg trending
  cg trending -o json`,
	RunE: runTrending,
}

const (
	maxTrendingCoins      = 15
	maxTrendingNFTs       = 7
	maxTrendingCategories = 6
)

func init() {
	rootCmd.AddCommand(trendingCmd)
}

func runTrending(cmd *cobra.Command, args []string) error {
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

	resp, err := client.SearchTrending(ctx)
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
		if i >= maxTrendingCoins {
			break
		}
		rank := "-"
		if c.Item.MarketCapRank > 0 {
			rank = fmt.Sprintf("%d", c.Item.MarketCapRank)
		}
		coinRows = append(coinRows, []string{
			fmt.Sprintf("%d", i+1),
			c.Item.Name,
			c.Item.Symbol,
			rank,
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
			if i >= maxTrendingNFTs {
				break
			}
			nftRows = append(nftRows, []string{
				fmt.Sprintf("%d", i+1),
				n.Name,
				n.Symbol,
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
			if i >= maxTrendingCategories {
				break
			}
			catRows = append(catRows, []string{
				fmt.Sprintf("%d", i+1),
				cat.Name,
				display.ColorPercent(cat.MarketCap1hChange),
			})
		}
		display.PrintTable(catHeaders, catRows)
	}

	return nil
}
