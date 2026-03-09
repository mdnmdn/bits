package cmd

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
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
  --days N              Price data for the last N days (or OHLC with --ohlc)
  --from/--to           Price data for a date range (or OHLC with --ohlc)

The --to date is inclusive: it covers the full day up to 23:59:59 UTC.
The --interval flag controls data granularity: daily or hourly.
Large ranges are automatically batched into multiple API requests.
The --ohlc flag switches --days and --from/--to to OHLC output (OHLC --days accepts: 1, 7, 14, 30, 90, 180, 365, max).`,
	Example: `  cg history bitcoin --date 2024-01-01
  cg history ethereum --days 30
  cg history bitcoin --days 7 --ohlc
  cg history bitcoin --days 90 --interval daily
  cg history solana --from 2024-01-01 --to 2024-03-01
  cg history solana --from 2024-01-01 --to 2024-03-01 --ohlc
  cg history solana --from 2024-01-01 --to 2024-03-01 --export prices.csv`,
	Args: cobra.ExactArgs(1),
	RunE: runHistory,
}

func init() {
	historyCmd.Flags().String("date", "", "Snapshot date (YYYY-MM-DD)")
	historyCmd.Flags().String("days", "", "Data for last N days (any integer, or max)")
	historyCmd.Flags().String("from", "", "Range start date (YYYY-MM-DD)")
	historyCmd.Flags().String("to", "", "Range end date (YYYY-MM-DD)")
	historyCmd.Flags().String("vs", "usd", "Target currency")
	historyCmd.Flags().String("interval", "", "Data interval: daily, hourly (auto-batched for large ranges)")
	historyCmd.Flags().Bool("ohlc", false, "Output OHLC data instead of price (for --days and --from/--to); --days only accepts 1,7,14,30,90,180,365,max in OHLC mode")
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
	ohlc, _ := cmd.Flags().GetBool("ohlc")
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

	if ohlc && dateStr != "" {
		return fmt.Errorf("--ohlc cannot be used with --date")
	}

	if ohlc && (fromStr != "" || toStr != "") && interval == "" {
		return fmt.Errorf("--ohlc with --from/--to requires --interval (daily or hourly)")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if ohlc && (fromStr != "" || toStr != "") && !cfg.IsPaid() {
		return fmt.Errorf("--ohlc with --from/--to: %w", api.ErrPlanRestricted)
	}

	if interval != "" && interval != "daily" && interval != "hourly" {
		return fmt.Errorf("invalid --interval %q — must be daily or hourly", interval)
	}

	if ohlc && interval != "" && !cfg.IsPaid() {
		return fmt.Errorf("--ohlc with --interval: %w", api.ErrPlanRestricted)
	}

	if ohlc && daysStr != "" && !validEnum("history", "days", daysStr) {
		return fmt.Errorf("invalid --days %q for OHLC — must be one of: 1, 7, 14, 30, 90, 180, 365, max", daysStr)
	}

	if isDryRun(cmd) {
		return historyDryRun(cfg, coinID, dateStr, daysStr, fromStr, toStr, vs, interval, ohlc)
	}

	client := api.NewClient(cfg)
	ctx := cmd.Context()

	switch {
	case dateStr != "":
		return historyDate(ctx, client, coinID, dateStr, vs, jsonOut)
	case daysStr != "":
		if ohlc {
			return historyOHLC(ctx, client, coinID, vs, daysStr, interval, exportPath, jsonOut)
		}
		return historyDays(ctx, client, coinID, vs, daysStr, interval, exportPath, jsonOut)
	default:
		if fromStr == "" || toStr == "" {
			return fmt.Errorf("both --from and --to are required for range mode")
		}
		if ohlc {
			return historyOHLCRange(ctx, client, coinID, vs, fromStr, toStr, interval, exportPath, jsonOut)
		}
		return historyRange(ctx, client, coinID, vs, fromStr, toStr, interval, exportPath, jsonOut)
	}
}

func historyDryRun(cfg *config.Config, coinID, dateStr, daysStr, fromStr, toStr, vs, interval string, ohlc bool) error {
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
		opKey := "--days"
		endpoint := "/coins/" + coinID + "/market_chart"
		note := ""
		if ohlc {
			opKey = "--days --ohlc"
			endpoint = "/coins/" + coinID + "/ohlc"
			if interval != "" {
				params["interval"] = interval
			}
		} else if interval == "hourly" && daysStr != "max" {
			// Real path routes hourly through /range with auto-granularity (no interval param).
			// Only rewrite if days is a valid integer; otherwise fall through to match live behavior.
			if n, err := strconv.Atoi(daysStr); err == nil {
				opKey = "--from/--to"
				endpoint = "/coins/" + coinID + "/market_chart/range"
				delete(params, "days")
				now := time.Now().UTC()
				apiDays := n
				if apiDays < minHourlyRangeDays {
					apiDays = minHourlyRangeDays
				}
				params["from"] = fmt.Sprintf("%d", now.AddDate(0, 0, -apiDays).Unix())
				params["to"] = fmt.Sprintf("%d", now.Unix())
				note = "hourly: batched via /range with auto-granularity (no interval param sent)"
			} else {
				params["interval"] = interval
			}
		} else if interval != "" {
			params["interval"] = interval
		}
		return printDryRunFull(cfg, "history", opKey, endpoint, params, nil, note)
	default:
		fromUnix, toUnix, err := parseRange(fromStr, toStr)
		if err != nil {
			return err
		}
		params := map[string]string{
			"vs_currency": vs,
			"from":        fmt.Sprintf("%d", fromUnix),
			"to":          fmt.Sprintf("%d", toUnix),
		}
		opKey := "--from/--to"
		endpoint := "/coins/" + coinID + "/market_chart/range"
		note := ""
		if ohlc {
			opKey = "--from/--to --ohlc"
			endpoint = "/coins/" + coinID + "/ohlc/range"
			if interval != "" {
				params["interval"] = interval
			}
		} else if interval == "hourly" {
			// Real path uses auto-granularity (no interval param sent).
			note = "hourly: batched via /range with auto-granularity (no interval param sent)"
		} else if interval == "daily" {
			// Real path uses auto-granularity; pads short ranges to 91 days.
			rangeDays := (toUnix - fromUnix) / secsPerDay
			if rangeDays < minDailyRangeDays {
				params["from"] = fmt.Sprintf("%d", toUnix-minDailyRangeDays*secsPerDay)
				note = "daily: range padded to 91 days for daily auto-granularity, results trimmed to original range"
			} else {
				note = "daily: uses auto-granularity (no interval param sent)"
			}
		}
		return printDryRunFull(cfg, "history", opKey, endpoint, params, nil, note)
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

func parseRange(fromStr, toStr string) (fromUnix, toUnix int64, err error) {
	fromTime, err := parseDate("--from", fromStr)
	if err != nil {
		return 0, 0, err
	}
	toTime, err := parseDate("--to", toStr)
	if err != nil {
		return 0, 0, err
	}
	return fromTime.UTC().Unix(), endOfDayUnix(toTime), nil
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

func historyDays(ctx context.Context, client *api.Client, coinID, vs, days, interval, exportPath string, jsonOut bool) error {
	// For hourly, always route through /range with auto-granularity.
	// Omit the interval param; ≤90-day chunks trigger hourly auto-granularity on all plans.
	// Auto-granularity requires ≥2 days for hourly; pad 1-day requests to 2 days then trim.
	if interval == "hourly" && days != "max" {
		n, err := strconv.Atoi(days)
		if err == nil {
			now := time.Now().UTC()
			apiDays := n
			if apiDays < minHourlyRangeDays {
				apiDays = minHourlyRangeDays
			}
			fromUnix := now.AddDate(0, 0, -apiDays).Unix()
			originalFrom := now.AddDate(0, 0, -n).Unix()
			toUnix := now.Unix()
			data, err := fetchMarketChartBatched(ctx, client, coinID, vs, fromUnix, toUnix, "", hourlyChunkDays)
			if err != nil {
				return err
			}
			if apiDays != n {
				fromMs := originalFrom * 1000
				data.Prices = trimTimeseries(data.Prices, fromMs)
				data.MarketCaps = trimTimeseries(data.MarketCaps, fromMs)
				data.TotalVolumes = trimTimeseries(data.TotalVolumes, fromMs)
			}
			if jsonOut {
				return printJSONRaw(data)
			}
			return renderPriceTable(data.Prices, vs, exportPath)
		}
	}

	// For daily and no-interval, /market_chart supports explicit interval=daily on all plans.
	data, err := client.CoinMarketChart(ctx, coinID, vs, days, interval)
	if err != nil {
		return err
	}

	if jsonOut {
		return printJSONRaw(data)
	}

	return renderPriceTable(data.Prices, vs, exportPath)
}

func historyRange(ctx context.Context, client *api.Client, coinID, vs, fromStr, toStr, interval, exportPath string, jsonOut bool) error {
	fromUnix, toUnix, err := parseRange(fromStr, toStr)
	if err != nil {
		return err
	}

	// For hourly, always batch with auto-granularity (omit interval param).
	// ≤90-day chunks trigger hourly auto-granularity on all plans.
	if interval == "hourly" {
		data, err := fetchMarketChartBatched(ctx, client, coinID, vs, fromUnix, toUnix, "", hourlyChunkDays)
		if err != nil {
			return err
		}
		if jsonOut {
			return printJSONRaw(data)
		}
		return renderPriceTable(data.Prices, vs, exportPath)
	}

	// For daily, use auto-granularity on /range (omit interval param).
	// Auto-granularity returns daily for ranges > 90 days.
	// If range is shorter, pad to 91 days then trim results.
	if interval == "daily" {
		apiFrom := fromUnix
		rangeDays := (toUnix - fromUnix) / secsPerDay
		if rangeDays < minDailyRangeDays {
			apiFrom = toUnix - minDailyRangeDays*secsPerDay
		}
		data, err := client.CoinMarketChartRange(ctx, coinID, vs, apiFrom, toUnix, "")
		if err != nil {
			return err
		}
		if apiFrom != fromUnix {
			fromMs := fromUnix * 1000
			data.Prices = trimTimeseries(data.Prices, fromMs)
			data.MarketCaps = trimTimeseries(data.MarketCaps, fromMs)
			data.TotalVolumes = trimTimeseries(data.TotalVolumes, fromMs)
		}
		if jsonOut {
			return printJSONRaw(data)
		}
		return renderPriceTable(data.Prices, vs, exportPath)
	}

	// No interval specified — pass through, let API auto-granularity handle it.
	data, err := client.CoinMarketChartRange(ctx, coinID, vs, fromUnix, toUnix, "")
	if err != nil {
		return err
	}

	if jsonOut {
		return printJSONRaw(data)
	}

	return renderPriceTable(data.Prices, vs, exportPath)
}

func historyOHLC(ctx context.Context, client *api.Client, coinID, vs, days, interval, exportPath string, jsonOut bool) error {
	data, err := client.CoinOHLC(ctx, coinID, vs, days, interval)
	if err != nil {
		return err
	}

	if jsonOut {
		return printJSONRaw(data)
	}

	return renderOHLCTable(data, vs, exportPath)
}

func historyOHLCRange(ctx context.Context, client *api.Client, coinID, vs, fromStr, toStr, interval, exportPath string, jsonOut bool) error {
	fromUnix, toUnix, err := parseRange(fromStr, toStr)
	if err != nil {
		return err
	}

	// Batch OHLC range requests when range exceeds per-request limits.
	// daily: 180 days max, hourly: 31 days max.
	chunkDays := ohlcRangeChunkDays(interval)
	rangeDays := (toUnix - fromUnix) / secsPerDay
	if chunkDays > 0 && rangeDays > int64(chunkDays) {
		data, err := fetchOHLCRangeBatched(ctx, client, coinID, vs, fromUnix, toUnix, interval, chunkDays)
		if err != nil {
			return err
		}
		if jsonOut {
			return printJSONRaw(data)
		}
		return renderOHLCTable(data, vs, exportPath)
	}

	data, err := client.CoinOHLCRange(ctx, coinID, vs, fromUnix, toUnix, interval)
	if err != nil {
		return err
	}

	if jsonOut {
		return printJSONRaw(data)
	}

	return renderOHLCTable(data, vs, exportPath)
}

const (
	secsPerDay          = 86400
	minHourlyRangeDays  = 2  // auto-granularity gives hourly starting at 2-day ranges
	hourlyChunkDays     = 90 // auto-granularity gives hourly for 2-90 day ranges
	minDailyRangeDays   = 91 // auto-granularity gives daily for >90 day ranges
)

// ohlcRangeChunkDays returns the max safe chunk size for OHLC range requests.
func ohlcRangeChunkDays(interval string) int {
	switch interval {
	case "daily":
		return 170 // API limit is 180 days; use 170 for safety
	case "hourly":
		return 30 // API limit is 31 days; use 30 for safety
	default:
		return 0
	}
}

// trimTimeseries removes data points with timestamps before fromMs (milliseconds).
// Timestamps are sorted ascending, so we binary search for the cut point and re-slice.
func trimTimeseries(data [][]float64, fromMs int64) [][]float64 {
	i := sort.Search(len(data), func(j int) bool {
		return len(data[j]) >= 2 && int64(data[j][0]) >= fromMs
	})
	return data[i:]
}

// dedupTimeseries appends entries from src to dst, skipping entries with timestamps
// already in seen. Each entry must have at least minLen elements; entry[0] is the timestamp.
func dedupTimeseries(dst *[][]float64, src [][]float64, seen map[int64]bool, minLen int) {
	for _, p := range src {
		if len(p) >= minLen {
			ts := int64(p[0])
			if !seen[ts] {
				seen[ts] = true
				*dst = append(*dst, p)
			}
		}
	}
}

const maxChunkRetries = 3

// retrySleep controls the sleep mechanism for rate-limit retries.
// Overridden in tests to avoid wall-clock delays.
var retrySleep = func(d time.Duration) <-chan time.Time { return time.After(d) }

// withRateLimitRetry retries fn on 429 responses, using Retry-After when available,
// then x-ratelimit-reset, then exponential backoff with jitter.
func withRateLimitRetry(ctx context.Context, chunkLabel string, fn func() error) error {
	var err error
	for attempt := 0; attempt <= maxChunkRetries; attempt++ {
		if attempt > 0 {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return ctxErr
			}
		}
		err = fn()
		if err == nil {
			return nil
		}
		var rle *api.RateLimitError
		if !errors.As(err, &rle) {
			return err // not rate limited, don't retry
		}
		if attempt == maxChunkRetries {
			break // retries exhausted
		}
		var wait time.Duration
		if rle.RetryAfter > 0 {
			// Server told us exactly how long to wait.
			wait = time.Duration(rle.RetryAfter) * time.Second
		} else if !rle.ResetAt.IsZero() {
			// Server told us when the rate limit resets.
			wait = time.Until(rle.ResetAt)
			if wait < time.Second {
				wait = time.Second // floor to avoid spinning
			}
		} else {
			// No server hint — exponential backoff with jitter: base * [0.5, 1.5)
			base := float64(int(2) << uint(attempt)) // 2, 4, 8
			jitter := 0.5 + rand.Float64()            // [0.5, 1.5)
			wait = time.Duration(base*jitter*1000) * time.Millisecond
		}
		if display.StderrIsTerminal() {
			warnf("\r  Rate limited (%s), retrying in %v...      \n", chunkLabel, wait.Round(time.Millisecond))
		} else {
			warnf("Rate limited (%s), retrying in %v...\n", chunkLabel, wait.Round(time.Millisecond))
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-retrySleep(wait):
		}
	}
	return err
}

// progressBar renders an inline progress indicator on stderr.
// On TTYs it overwrites the line in-place; off-TTY it prints simple log lines.
func progressBar(current, total int) {
	if !display.StderrIsTerminal() {
		if current == 0 {
			warnf("Fetching data")
		}
		if current < total {
			warnf(".")
		} else {
			warnf("\n")
		}
		return
	}
	const barWidth = 20
	filled := barWidth * current / total
	var filledBar, emptyBar string
	for range filled {
		filledBar += "█"
	}
	for range barWidth - filled {
		emptyBar += "░"
	}
	pct := 100 * current / total
	if display.ColorEnabled() {
		// Brand green #4BCC00 for the filled portion only.
		fmt.Fprintf(os.Stderr, "\r  Fetching data... \033[38;2;75;204;0m%s\033[0m%s %d%% (%d/%d)", filledBar, emptyBar, pct, current, total)
	} else {
		fmt.Fprintf(os.Stderr, "\r  Fetching data... %s%s %d%% (%d/%d)", filledBar, emptyBar, pct, current, total)
	}
	if current == total {
		fmt.Fprint(os.Stderr, "\n")
	}
}

// progressClear terminates the in-place progress line so subsequent
// output (e.g. error messages) starts on a fresh line.
func progressClear() {
	fmt.Fprint(os.Stderr, "\n")
}

// fetchOHLCRangeBatched splits a large OHLC date range into chunks and merges results.
func fetchOHLCRangeBatched(ctx context.Context, client *api.Client, coinID, vs string, fromUnix, toUnix int64, interval string, chunkDays int) (api.OHLCData, error) {
	// Cap toUnix at current time — the OHLC range endpoint rejects future dates.
	if now := time.Now().Unix(); toUnix > now {
		toUnix = now
	}
	chunkSec := int64(chunkDays) * secsPerDay
	totalChunks := int((toUnix - fromUnix + chunkSec - 1) / chunkSec)

	var merged api.OHLCData
	seen := make(map[int64]bool)

	for i := 0; i < totalChunks; i++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		chunkFrom := fromUnix + int64(i)*chunkSec
		chunkTo := chunkFrom + chunkSec
		if chunkTo > toUnix {
			chunkTo = toUnix
		}

		chunkLabel := fmt.Sprintf("chunk %d/%d", i+1, totalChunks)
		progressBar(i, totalChunks)

		var data api.OHLCData
		err := withRateLimitRetry(ctx, chunkLabel, func() error {
			var fetchErr error
			data, fetchErr = client.CoinOHLCRange(ctx, coinID, vs, chunkFrom, chunkTo, interval)
			return fetchErr
		})
		if err != nil {
			progressClear()
			return nil, fmt.Errorf("chunk %d/%d failed: %w", i+1, totalChunks, err)
		}

		dedupTimeseries((*[][]float64)(&merged), [][]float64(data), seen, 5)
	}
	progressBar(totalChunks, totalChunks)

	return merged, nil
}

// fetchMarketChartBatched splits a large date range into chunks and merges results.
func fetchMarketChartBatched(ctx context.Context, client *api.Client, coinID, vs string, fromUnix, toUnix int64, interval string, chunkDays int) (*api.MarketChartResponse, error) {
	// Cap toUnix at current time — some endpoints reject future dates.
	if now := time.Now().Unix(); toUnix > now {
		toUnix = now
	}
	chunkSec := int64(chunkDays) * secsPerDay
	// Calculate number of chunks.
	totalChunks := int((toUnix-fromUnix+chunkSec-1) / chunkSec)

	merged := &api.MarketChartResponse{}
	seenPrices := make(map[int64]bool)
	seenMCaps := make(map[int64]bool)
	seenVols := make(map[int64]bool)

	for i := 0; i < totalChunks; i++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		chunkFrom := fromUnix + int64(i)*chunkSec
		chunkTo := chunkFrom + chunkSec
		if chunkTo > toUnix {
			chunkTo = toUnix
		}

		chunkLabel := fmt.Sprintf("chunk %d/%d", i+1, totalChunks)
		progressBar(i, totalChunks)

		var data *api.MarketChartResponse
		err := withRateLimitRetry(ctx, chunkLabel, func() error {
			var fetchErr error
			data, fetchErr = client.CoinMarketChartRange(ctx, coinID, vs, chunkFrom, chunkTo, interval)
			return fetchErr
		})
		if err != nil {
			progressClear()
			return nil, fmt.Errorf("chunk %d/%d failed: %w", i+1, totalChunks, err)
		}

		dedupTimeseries(&merged.Prices, data.Prices, seenPrices, 2)
		dedupTimeseries(&merged.MarketCaps, data.MarketCaps, seenMCaps, 2)
		dedupTimeseries(&merged.TotalVolumes, data.TotalVolumes, seenVols, 2)
	}
	progressBar(totalChunks, totalChunks)

	return merged, nil
}

func renderPriceTable(prices [][]float64, vs, exportPath string) error {
	headers := []string{"Date", "Price"}
	rows := make([][]string, 0, len(prices))
	var csvRows [][]string
	if exportPath != "" {
		csvRows = make([][]string, 0, len(prices))
	}
	for _, p := range prices {
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

func renderOHLCTable(data api.OHLCData, vs, exportPath string) error {
	headers := []string{"Date", "Open", "High", "Low", "Close"}
	rows := make([][]string, 0, len(data))
	var csvRows [][]string
	if exportPath != "" {
		csvRows = make([][]string, 0, len(data))
	}
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
