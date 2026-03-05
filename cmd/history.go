package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"coingecko-cli/internal/api"
	"coingecko-cli/internal/config"
	"coingecko-cli/internal/display"

	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history [coin-id]",
	Short: "Get historical price data for a coin",
	Long: `Fetch historical data using one of three modes (mutually exclusive):
  --date YYYY-MM-DD     Snapshot on a specific date
  --days N              OHLC data for the last N days
  --from/--to           Price data for a date range (YYYY-MM-DD)

The --to date is inclusive: it covers the full day up to 23:59:59 UTC.`,
	Example: `  cg history bitcoin --date 2024-01-01
  cg history ethereum --days 30
  cg history solana --from 2024-01-01 --to 2024-03-01 --export prices.csv`,
	Args: cobra.ExactArgs(1),
	RunE: runHistory,
}

func init() {
	historyCmd.Flags().String("date", "", "Snapshot date (YYYY-MM-DD)")
	historyCmd.Flags().Int("days", 0, "OHLC data for last N days")
	historyCmd.Flags().String("from", "", "Range start date (YYYY-MM-DD)")
	historyCmd.Flags().String("to", "", "Range end date (YYYY-MM-DD)")
	historyCmd.Flags().String("vs", "usd", "Target currency")
	historyCmd.Flags().String("export", "", "Export to CSV file path")
	rootCmd.AddCommand(historyCmd)
}

func runHistory(cmd *cobra.Command, args []string) error {
	coinID := args[0]
	dateStr, _ := cmd.Flags().GetString("date")
	days, _ := cmd.Flags().GetInt("days")
	fromStr, _ := cmd.Flags().GetString("from")
	toStr, _ := cmd.Flags().GetString("to")
	vs, _ := cmd.Flags().GetString("vs")
	exportPath, _ := cmd.Flags().GetString("export")
	jsonOut := outputJSON(cmd)

	if !jsonOut {
		display.PrintBanner()
	}

	modes := 0
	if dateStr != "" {
		modes++
	}
	if days > 0 {
		modes++
	}
	if fromStr != "" || toStr != "" {
		modes++
	}
	if modes != 1 {
		return fmt.Errorf("specify exactly one mode: --date, --days, or --from/--to")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	client := api.NewClient(cfg)
	ctx := cmd.Context()

	switch {
	case dateStr != "":
		return historyDate(ctx, client, coinID, dateStr, vs, jsonOut)
	case days > 0:
		return historyOHLC(ctx, client, coinID, vs, days, exportPath, jsonOut)
	default:
		return historyRange(ctx, client, coinID, vs, fromStr, toStr, exportPath, jsonOut)
	}
}

func historyDate(ctx context.Context, client *api.Client, coinID, dateStr, vs string, jsonOut bool) error {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid date format, use YYYY-MM-DD: %w", err)
	}
	apiDate := t.Format("02-01-2006") // CoinGecko uses DD-MM-YYYY

	data, err := client.CoinHistory(ctx, coinID, apiDate)
	if err != nil {
		return err
	}

	if data.MarketData == nil {
		return fmt.Errorf("no market data available for %s on %s", coinID, dateStr)
	}

	if jsonOut {
		return printJSONRaw(data)
	}

	headers := []string{"Metric", "Value"}
	rows := [][]string{
		{"Coin", fmt.Sprintf("%s (%s)", data.Name, strings.ToUpper(data.Symbol))},
		{"Date", dateStr},
		{"Price", display.FormatPrice(data.MarketData.CurrentPrice[vs])},
		{"Market Cap", display.FormatLargeNumber(data.MarketData.MarketCap[vs])},
		{"Volume", display.FormatLargeNumber(data.MarketData.TotalVolume[vs])},
	}
	display.PrintTable(headers, rows)
	return nil
}

func historyOHLC(ctx context.Context, client *api.Client, coinID, vs string, days int, exportPath string, jsonOut bool) error {
	data, err := client.CoinOHLC(ctx, coinID, vs, days)
	if err != nil {
		return err
	}

	if jsonOut {
		return printJSONRaw(data)
	}

	headers := []string{"Date", "Open", "High", "Low", "Close"}
	var rows [][]string
	for _, d := range data {
		if len(d) < 5 {
			continue
		}
		ts := time.UnixMilli(int64(d[0]))
		rows = append(rows, []string{
			ts.UTC().Format("2006-01-02 15:04"),
			display.FormatPrice(d[1]),
			display.FormatPrice(d[2]),
			display.FormatPrice(d[3]),
			display.FormatPrice(d[4]),
		})
	}

	display.PrintTable(headers, rows)

	if exportPath != "" {
		var csvRows [][]string
		for _, d := range data {
			if len(d) < 5 {
				continue
			}
			ts := time.UnixMilli(int64(d[0]))
			csvRows = append(csvRows, []string{
				ts.UTC().Format(time.RFC3339),
				fmt.Sprintf("%.8f", d[1]),
				fmt.Sprintf("%.8f", d[2]),
				fmt.Sprintf("%.8f", d[3]),
				fmt.Sprintf("%.8f", d[4]),
			})
		}
		if err := exportCSV(exportPath, headers, csvRows); err != nil {
			return err
		}
	}
	return nil
}

func historyRange(ctx context.Context, client *api.Client, coinID, vs, fromStr, toStr, exportPath string, jsonOut bool) error {
	if fromStr == "" || toStr == "" {
		return fmt.Errorf("both --from and --to are required for range mode")
	}

	fromTime, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		return fmt.Errorf("invalid --from date, use YYYY-MM-DD: %w", err)
	}
	toTime, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		return fmt.Errorf("invalid --to date, use YYYY-MM-DD: %w", err)
	}

	from := fromTime.UTC().Unix()
	to := toTime.Add(23*time.Hour + 59*time.Minute + 59*time.Second).UTC().Unix()

	data, err := client.CoinMarketChartRange(ctx, coinID, vs, from, to)
	if err != nil {
		return err
	}

	if jsonOut {
		return printJSONRaw(data)
	}

	headers := []string{"Date", "Price"}
	var rows [][]string
	for _, p := range data.Prices {
		if len(p) < 2 {
			continue
		}
		ts := time.UnixMilli(int64(p[0]))
		rows = append(rows, []string{
			ts.UTC().Format("2006-01-02 15:04"),
			display.FormatPrice(p[1]),
		})
	}

	display.PrintTable(headers, rows)

	if exportPath != "" {
		var csvRows [][]string
		for _, p := range data.Prices {
			if len(p) < 2 {
				continue
			}
			ts := time.UnixMilli(int64(p[0]))
			csvRows = append(csvRows, []string{
				ts.UTC().Format(time.RFC3339),
				fmt.Sprintf("%.8f", p[1]),
			})
		}
		if err := exportCSV(exportPath, headers, csvRows); err != nil {
			return err
		}
	}
	return nil
}
