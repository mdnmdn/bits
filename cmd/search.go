package cmd

import (
	"context"
	"fmt"

	"coingecko-cli/internal/api"
	"coingecko-cli/internal/config"
	"coingecko-cli/internal/display"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for coins",
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

func init() {
	searchCmd.Flags().Int("limit", 10, "Max results to show")
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	limit, _ := cmd.Flags().GetInt("limit")
	if limit < 0 {
		limit = 0
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	client := api.NewClient(cfg)
	ctx := context.Background()

	resp, err := client.Search(ctx, args[0])
	if err != nil {
		return err
	}

	coins := resp.Coins
	if len(coins) > limit {
		coins = coins[:limit]
	}

	headers := []string{"Rank", "Name", "Symbol", "ID"}
	rows := make([][]string, len(coins))
	for i, c := range coins {
		rank := "-"
		if c.MarketCapRank > 0 {
			rank = fmt.Sprintf("%d", c.MarketCapRank)
		}
		rows[i] = []string{rank, c.Name, c.Symbol, c.ID}
	}

	display.PrintTable(headers, rows)
	return nil
}
