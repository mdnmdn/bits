package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/coingecko/coingecko-cli/internal/api"
	"github.com/coingecko/coingecko-cli/internal/config"
	"github.com/coingecko/coingecko-cli/internal/display"

	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history [coin-id]",
	Short: "Get historical price data for a coin",
	Long: `Fetch historical data using one of three modes (mutually exclusive):
  --date YYYY-MM-DD     Snapshot on a specific date
  --days N              OHLC data for the last N days
  --from/--to           Price data for a date range (YYYY-MM-DD)

The --to date is inclusive: it covers the full day up to 23:59:59 UTC.
The --interval flag (paid plans only) controls candle/data granularity.`,
	Example: `  cg history bitcoin --date 2024-01-01
  cg history ethereum --days 30
  cg history bitcoin --days 90 --interval daily
  cg history solana --from 2024-01-01 --to 2024-03-01 --interval hourly
  cg history solana --from 2024-01-01 --to 2024-03-01 --export prices.csv`,
	Args: cobra.ExactArgs(1),
	RunE: runHistory,
}

func init() {
	historyCmd.Flags().String("date", "", "Snapshot date (YYYY-MM-DD)")
	historyCmd.Flags().String("days", "", "OHLC data for last N days (1, 7, 14, 30, 90, 180, 365, max)")
	historyCmd.Flags().String("from", "", "Range start date (YYYY-MM-DD)")
	historyCmd.Flags().String("to", "", "Range end date (YYYY-MM-DD)")
	historyCmd.Flags().String("vs", "usd", "Target currency")
	historyCmd.Flags().String("interval", "", "Data interval: daily, hourly (paid plans only); 5m for range (Enterprise)")
	historyCmd.Flags().String("export", "", "Export to CSV file path")
	rootCmd.AddCommand(historyCmd)
}

func runHistory(cmd *cobra.Command, args []string) error {
	coinID := args[0]
	dateStr, _ := cmd.Flags().GetString("date")
	daysStr, _ := cmd.Flags().GetString("days")
	fromStr, _ := cmd.Flags().GetString("from")
	toStr, _ := cmd.Flags().GetString("to")
	vs, _ := cmd.Flags().GetString("vs")
	interval, _ := cmd.Flags().GetString("interval")
	exportPath, _ := cmd.Flags().GetString("export")
	jsonOut := outputJSON(cmd)

	if !jsonOut {
		display.PrintBanner()
	}

	modes := 0
	if dateStr != "" {
		modes++
	}
	if daysStr != "" {
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

	if interval != "" && !cfg.IsPaid() {
		return fmt.Errorf("--interval requires a paid plan — upgrade at %s", paidPlanURL)
	}

	if isDryRun(cmd) {
		switch {
		case dateStr != "":
			t, err := parseDate("--date", dateStr)
			if err != nil {
				return err
			}
			return printDryRunWithOp(cfg, "history", "--date", "/coins/"+coinID+"/history", map[string]string{
				"date": t.Format("02-01-2006"), "localization": "false",
			}, nil)
		case daysStr != "":
			params := map[string]string{"vs_currency": vs, "days": daysStr}
			if interval != "" {
				params["interval"] = interval
			}
			return printDryRunWithOp(cfg, "history", "--days", "/coins/"+coinID+"/ohlc", params, nil)
		default:
			if fromStr == "" || toStr == "" {
				return fmt.Errorf("both --from and --to are required for range mode")
			}
			fromTime, err := parseDate("--from", fromStr)
			if err != nil {
				return err
			}
			toTime, err := parseDate("--to", toStr)
			if err != nil {
				return err
			}
			params := map[string]string{
				"vs_currency": vs,
				"from":        fmt.Sprintf("%d", fromTime.UTC().Unix()),
				"to":          fmt.Sprintf("%d", endOfDayUnix(toTime)),
			}
			if interval != "" {
				params["interval"] = interval
			}
			return printDryRunWithOp(cfg, "history", "--from/--to", "/coins/"+coinID+"/market_chart/range", params, nil)
		}
	}

	client := api.NewClient(cfg)
	ctx := cmd.Context()

	switch {
	case dateStr != "":
		return historyDate(ctx, client, coinID, dateStr, vs, jsonOut)
	case daysStr != "":
		if !validEnum("history", "days", daysStr) {
			return fmt.Errorf("invalid --days %q — must be one of: 1, 7, 14, 30, 90, 180, 365, max", daysStr)
		}
		if daysStr == "max" && !cfg.IsPaid() {
			return fmt.Errorf("--days max requires a paid plan — upgrade at %s", paidPlanURL)
		}
		return historyOHLC(ctx, client, coinID, vs, daysStr, interval, exportPath, jsonOut)
	default:
		return historyRange(ctx, client, coinID, vs, fromStr, toStr, interval, exportPath, jsonOut)
	}
}

const dateLayout = "2006-01-02"

func parseDate(name, value string) (time.Time, error) {
	t, err := time.Parse(dateLayout, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid %s date, use YYYY-MM-DD: %w", name, err)
	}
	return t, nil
}

func endOfDayUnix(t time.Time) int64 {
	return t.Add(23*time.Hour + 59*time.Minute + 59*time.Second).UTC().Unix()
}

func historyDate(ctx context.Context, client *api.Client, coinID, dateStr, vs string, jsonOut bool) error {
	t, err := parseDate("--date", dateStr)
	if err != nil {
		return err
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
		{"Coin", fmt.Sprintf("%s (%s)", display.SanitizeCell(data.Name), display.FormatSymbol(data.Symbol))},
		{"Date", dateStr},
		{"Price", display.FormatPrice(data.MarketData.CurrentPrice[vs], vs)},
		{"Market Cap", display.FormatLargeNumber(data.MarketData.MarketCap[vs], vs)},
		{"Volume", display.FormatLargeNumber(data.MarketData.TotalVolume[vs], vs)},
	}
	display.PrintTable(headers, rows)
	return nil
}

func historyOHLC(ctx context.Context, client *api.Client, coinID, vs, days, interval, exportPath string, jsonOut bool) error {
	data, err := client.CoinOHLC(ctx, coinID, vs, days, interval)
	if err != nil {
		return err
	}

	if jsonOut {
		return printJSONRaw(data)
	}

	headers := []string{"Date", "Open", "High", "Low", "Close"}
	var rows [][]string
	var csvRows [][]string
	for _, d := range data {
		if len(d) < 5 {
			continue
		}
		ts := time.UnixMilli(int64(d[0]))
		rows = append(rows, []string{
			ts.UTC().Format("2006-01-02 15:04"),
			display.FormatPrice(d[1], vs),
			display.FormatPrice(d[2], vs),
			display.FormatPrice(d[3], vs),
			display.FormatPrice(d[4], vs),
		})
		if exportPath != "" {
			csvRows = append(csvRows, []string{
				ts.UTC().Format(time.RFC3339),
				fmt.Sprintf("%.8f", d[1]),
				fmt.Sprintf("%.8f", d[2]),
				fmt.Sprintf("%.8f", d[3]),
				fmt.Sprintf("%.8f", d[4]),
			})
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

func historyRange(ctx context.Context, client *api.Client, coinID, vs, fromStr, toStr, interval, exportPath string, jsonOut bool) error {
	if fromStr == "" || toStr == "" {
		return fmt.Errorf("both --from and --to are required for range mode")
	}

	fromTime, err := parseDate("--from", fromStr)
	if err != nil {
		return err
	}
	toTime, err := parseDate("--to", toStr)
	if err != nil {
		return err
	}

	from := fromTime.UTC().Unix()
	to := endOfDayUnix(toTime)

	data, err := client.CoinMarketChartRange(ctx, coinID, vs, from, to, interval)
	if err != nil {
		return err
	}

	if jsonOut {
		return printJSONRaw(data)
	}

	headers := []string{"Date", "Price"}
	var rows [][]string
	var csvRows [][]string
	for _, p := range data.Prices {
		if len(p) < 2 {
			continue
		}
		ts := time.UnixMilli(int64(p[0]))
		rows = append(rows, []string{
			ts.UTC().Format("2006-01-02 15:04"),
			display.FormatPrice(p[1], vs),
		})
		if exportPath != "" {
			csvRows = append(csvRows, []string{
				ts.UTC().Format(time.RFC3339),
				fmt.Sprintf("%.8f", p[1]),
			})
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
