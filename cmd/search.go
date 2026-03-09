package cmd

import (
	"github.com/coingecko/coingecko-cli/internal/display"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for coins",
	Example: `  cg search bitcoin
  cg search sol --limit 5`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
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
	jsonOut := outputJSON(cmd)

	if !jsonOut {
		display.PrintBanner()
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	if isDryRun(cmd) {
		return printDryRun(cfg, "search", "/search", map[string]string{
			"query": args[0],
		}, nil)
	}

	client := newAPIClient(cfg)
	ctx := cmd.Context()

	resp, err := client.Search(ctx, args[0])
	if err != nil {
		return err
	}

	coins := resp.Coins
	if len(coins) > limit {
		coins = coins[:limit]
	}

	if jsonOut {
		return printJSONRaw(coins)
	}

	headers := []string{"Rank", "Name", "Symbol", "ID"}
	rows := make([][]string, len(coins))
	for i, c := range coins {
		rows[i] = []string{display.FormatRank(c.MarketCapRank), display.SanitizeCell(c.Name), display.SanitizeCell(c.Symbol), display.SanitizeCell(c.ID)}
	}

	display.PrintTable(headers, rows)
	return nil
}
