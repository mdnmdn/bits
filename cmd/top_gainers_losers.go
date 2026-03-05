package cmd

import (
	"fmt"
	"strings"

	"github.com/coingecko/coingecko-cli/internal/api"
	"github.com/coingecko/coingecko-cli/internal/config"
	"github.com/coingecko/coingecko-cli/internal/display"

	"github.com/spf13/cobra"
)

var topGainersLosersCmd = &cobra.Command{
	Use:   "top-gainers-losers",
	Short: "Show top gaining and losing coins (paid plans only)",
	Example: `  cg top-gainers-losers
  cg top-gainers-losers --losers --duration 7d
  cg top-gainers-losers --price-change-percentage 1h,7d,30d
  cg top-gainers-losers --top-coins 300 --export gainers.csv`,
	RunE: runTopGainersLosers,
}

const maxGainersLosersDisplay = 30

func init() {
	topGainersLosersCmd.Flags().String("vs", "usd", "Target currency")
	topGainersLosersCmd.Flags().String("duration", "24h", "Duration (1h, 24h, 7d, 14d, 30d, 60d, 1y)")
	topGainersLosersCmd.Flags().String("top-coins", "1000", "Top N coins by market cap (300, 500, 1000, all)")
	topGainersLosersCmd.Flags().String("price-change-percentage", "", "Include extra change %: 1h, 24h, 7d, 14d, 30d, 200d, 1y (comma-separated)")
	topGainersLosersCmd.Flags().Bool("losers", false, "Show losers instead of gainers")
	topGainersLosersCmd.Flags().String("export", "", "Export to CSV file path")
	rootCmd.AddCommand(topGainersLosersCmd)
}

func runTopGainersLosers(cmd *cobra.Command, args []string) error {
	vs, _ := cmd.Flags().GetString("vs")
	duration, _ := cmd.Flags().GetString("duration")
	topCoinsStr, _ := cmd.Flags().GetString("top-coins")
	priceChangePct, _ := cmd.Flags().GetString("price-change-percentage")
	showLosers, _ := cmd.Flags().GetBool("losers")
	exportPath, _ := cmd.Flags().GetString("export")
	jsonOut := outputJSON(cmd)

	if !jsonOut {
		display.PrintBanner()
	}

	validDurations := map[string]bool{"1h": true, "24h": true, "7d": true, "14d": true, "30d": true, "60d": true, "1y": true}
	if !validDurations[duration] {
		return fmt.Errorf("invalid duration %q — must be one of: 1h, 24h, 7d, 14d, 30d, 60d, 1y", duration)
	}

	validTopCoins := map[string]bool{"300": true, "500": true, "1000": true, "all": true}
	if !validTopCoins[topCoinsStr] {
		return fmt.Errorf("invalid --top-coins %q — must be 300, 500, 1000, or all", topCoinsStr)
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	client := api.NewClient(cfg)
	ctx := cmd.Context()

	resp, err := client.TopGainersLosers(ctx, vs, duration, topCoinsStr, priceChangePct)
	if err != nil {
		return err
	}

	if jsonOut {
		return printJSONRaw(resp)
	}

	coins := resp.TopGainers
	title := "Top Gainers"
	if showLosers {
		coins = resp.TopLosers
		title = "Top Losers"
	}

	fmt.Printf("%s (%s, top %s coins, vs %s)\n\n", title, duration, topCoinsStr, vs)

	headers := []string{"#", "Name", "Symbol", "Price", "Change %"}
	rows := make([][]string, 0, len(coins))
	for i := range coins {
		if i >= maxGainersLosersDisplay {
			break
		}
		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			display.SanitizeCell(coins[i].Name),
			strings.ToUpper(display.SanitizeCell(coins[i].Symbol)),
			display.FormatPrice(coins[i].Price(vs)),
			display.ColorPercent(coins[i].PriceChange(vs)),
		})
	}

	display.PrintTable(headers, rows)

	if exportPath != "" {
		csvRows := make([][]string, 0, len(rows))
		for i := range coins {
			if i >= maxGainersLosersDisplay {
				break
			}
			csvRows = append(csvRows, []string{
				fmt.Sprintf("%d", i+1),
				display.SanitizeCell(coins[i].Name),
				strings.ToUpper(display.SanitizeCell(coins[i].Symbol)),
				fmt.Sprintf("%.8f", coins[i].Price(vs)),
				fmt.Sprintf("%.2f", coins[i].PriceChange(vs)),
			})
		}
		if err := exportCSV(exportPath, headers, csvRows); err != nil {
			return err
		}
	}

	return nil
}
