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

var topGainersLosersCmd = &cobra.Command{
	Use:   "top-gainers-losers",
	Short: "Show top gaining and losing coins (paid plans only)",
	RunE:  runTopGainersLosers,
}

func init() {
	topGainersLosersCmd.Flags().String("vs", "usd", "Target currency")
	topGainersLosersCmd.Flags().String("duration", "24h", "Duration (1h, 24h, 7d, 14d, 30d, 60d, 1y)")
	topGainersLosersCmd.Flags().Int("top-coins", 1000, "Top N coins by market cap (300, 500, 1000)")
	topGainersLosersCmd.Flags().Bool("losers", false, "Show losers instead of gainers")
	topGainersLosersCmd.Flags().String("export", "", "Export to CSV file path")
	rootCmd.AddCommand(topGainersLosersCmd)
}

func runTopGainersLosers(cmd *cobra.Command, args []string) error {
	vs, _ := cmd.Flags().GetString("vs")
	duration, _ := cmd.Flags().GetString("duration")
	topCoins, _ := cmd.Flags().GetInt("top-coins")
	showLosers, _ := cmd.Flags().GetBool("losers")
	exportPath, _ := cmd.Flags().GetString("export")

	validDurations := map[string]bool{"1h": true, "24h": true, "7d": true, "14d": true, "30d": true, "60d": true, "1y": true}
	if !validDurations[duration] {
		return fmt.Errorf("invalid duration %q — must be one of: 1h, 24h, 7d, 14d, 30d, 60d, 1y", duration)
	}

	validTopCoins := map[int]bool{300: true, 500: true, 1000: true}
	if !validTopCoins[topCoins] {
		return fmt.Errorf("invalid --top-coins %d — must be 300, 500, or 1000", topCoins)
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	client := api.NewClient(cfg)
	ctx := context.Background()

	resp, err := client.TopGainersLosers(ctx, vs, duration, topCoins)
	if err != nil {
		return err
	}

	coins := resp.TopGainers
	title := "Top Gainers"
	if showLosers {
		coins = resp.TopLosers
		title = "Top Losers"
	}

	fmt.Printf("%s (%s, top %d coins)\n\n", title, duration, topCoins)

	headers := []string{"#", "Name", "Symbol", "Price", "Change %"}
	rows := make([][]string, 0, len(coins))
	for i, c := range coins {
		if i >= 30 {
			break
		}
		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			c.Name,
			strings.ToUpper(c.Symbol),
			display.FormatPrice(c.USD),
			display.ColorPercent(c.USDPriceChangePercentage),
		})
	}

	display.PrintTable(headers, rows)

	if exportPath != "" {
		csvRows := make([][]string, len(rows))
		for i, c := range coins {
			if i >= 30 {
				break
			}
			csvRows[i] = []string{
				fmt.Sprintf("%d", i+1),
				c.Name,
				strings.ToUpper(c.Symbol),
				fmt.Sprintf("%.8f", c.USD),
				fmt.Sprintf("%.2f", c.USDPriceChangePercentage),
			}
		}
		if err := export.ExportCSV(exportPath, headers, csvRows); err != nil {
			return err
		}
		fmt.Printf("Exported to %s\n", exportPath)
	}

	return nil
}
