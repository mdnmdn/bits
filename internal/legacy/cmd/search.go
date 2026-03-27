package legacycmd

import (
	"fmt"

	"github.com/mdnmdn/bits/internal/legacy/display"
	"github.com/mdnmdn/bits/internal/legacy/provider"

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

	client, err := newAPIClientWithFeature("search")(cfg)
	if err != nil {
		return err
	}
	ctx := cmd.Context()

	s, ok := client.(provider.Searcher)
	if !ok {
		return fmt.Errorf("%s provider does not support search", client.ID())
	}
	resp, err := s.Search(ctx, args[0])
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
