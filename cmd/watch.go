package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/coingecko/coingecko-cli/internal/display"
	"github.com/coingecko/coingecko-cli/internal/provider"
	"github.com/coingecko/coingecko-cli/internal/provider/coingecko"
	"github.com/coingecko/coingecko-cli/internal/ws"

	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Stream live coin prices via WebSocket (analyst or above)",
	Long:  "Connect to CoinGecko's real-time WebSocket API and stream live price updates. Requires an analyst plan or above.",
	Example: `  cg watch --ids bitcoin,ethereum
  cg watch --symbols btc,eth
  cg watch --ids bitcoin -o json
  cg watch --ids bitcoin --dry-run`,
	RunE: runWatch,
}

func init() {
	watchCmd.Flags().String("ids", "", "Comma-separated coin IDs (e.g. bitcoin,ethereum)")
	watchCmd.Flags().String("symbols", "", "Comma-separated symbols (e.g. btc,eth)")
	rootCmd.AddCommand(watchCmd)
}

func runWatch(cmd *cobra.Command, args []string) error {
	idsStr, _ := cmd.Flags().GetString("ids")
	symbolsStr, _ := cmd.Flags().GetString("symbols")
	jsonOut := outputJSON(cmd)

	if !jsonOut {
		display.PrintBanner()
	}

	if idsStr == "" && symbolsStr == "" {
		return fmt.Errorf("specify coins to watch with --ids or --symbols")
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if !cfg.IsPaid() {
		return coingecko.ErrPlanRestricted
	}

	var coinIDs []string
	var preflights []dryRunOutput
	dryRun := isDryRun(cmd)

	// Create API client once for both ID validation and symbol resolution.
	var client provider.Provider
	if !dryRun {
		client = newAPIClient(cfg)
	}

	if idsStr != "" {
		requested := splitTrim(idsStr)

		if dryRun {
			coinIDs = append(coinIDs, requested...)
			headerKey, _ := cfg.AuthHeader()
			preflights = append(preflights, dryRunOutput{
				Method: "GET",
				URL:    cfg.BaseURL() + "/simple/price",
				Params: map[string]string{
					"ids":            idsStr,
					"vs_currencies":  "usd",
					"include_24hr_change": "true",
				},
				Headers: map[string]string{
					headerKey: cfg.MaskedKey(),
					"Accept":  "application/json",
				},
				Note: "Validates coin IDs before connecting",
			})
		} else {
			prices, err := client.SimplePrice(cmd.Context(), requested, "usd")
			if err != nil {
				return fmt.Errorf("validating coin IDs: %w", err)
			}
			for _, id := range requested {
				if _, ok := prices[id]; ok {
					coinIDs = append(coinIDs, id)
				} else {
					warnf("coin ID %q not found (use --symbols for ticker symbols like btc), skipping\n", id)
				}
			}
		}
	}

	if symbolsStr != "" {
		symbols := splitTrim(symbolsStr)
		if dryRun {
			for _, sym := range symbols {
				headerKey, _ := cfg.AuthHeader()
				preflights = append(preflights, dryRunOutput{
					Method: "GET",
					URL:    cfg.BaseURL() + "/search",
					Params: map[string]string{
						"query": sym,
					},
					Headers: map[string]string{
						headerKey: cfg.MaskedKey(),
						"Accept":  "application/json",
					},
					Note: fmt.Sprintf("Resolves symbol %q to coin ID", sym),
				})
			}
		} else {
			for _, sym := range symbols {
				res, err := client.Search(cmd.Context(), sym)
				if err != nil {
					return fmt.Errorf("resolving symbol %q: %w", sym, err)
				}
				if id := matchSymbol(res.Coins, sym); id != "" {
					coinIDs = append(coinIDs, id)
				} else {
					warnf("could not resolve symbol %q to a coin, skipping\n", sym)
				}
			}
		}
	}

	if dryRun {
		return printDryRunWS(cfg, coinIDs, preflights)
	}

	if len(coinIDs) == 0 {
		return fmt.Errorf("none of the provided coins could be found — verify your --ids (e.g. bitcoin) or --symbols (e.g. btc)")
	}

	// Set up context with signal handling.
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-ctx.Done():
		}
		signal.Stop(sigCh)
	}()

	streamer := newStreamer(cfg, coinIDs)

	updates, err := streamer.Connect(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = streamer.Close() }()

	if jsonOut {
		if err := watchJSON(ctx, updates); err != nil {
			if isBrokenPipe(err) {
				cancel()
				return nil
			}
			return err
		}
		return nil
	}

	return watchTable(ctx, updates)
}

func watchJSON(ctx context.Context, updates <-chan *ws.CoinUpdate) error {
	enc := json.NewEncoder(os.Stdout)

	for {
		select {
		case update, ok := <-updates:
			if !ok {
				return nil
			}
			if err := enc.Encode(update); err != nil {
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}
}

// priceFlash tracks a temporary color flash for a price change.
type priceFlash struct {
	dir     int // +1 up, -1 down
	expires time.Time
}

const flashDuration = 500 * time.Millisecond

// statusDots cycles through ["   ", ".  ", ".. ", "..."] to show liveness.
var statusDots = [4]string{"   ", ".  ", ".. ", "..."}

// statusLine is the fixed-width status text template. The dots region is
// always 3 characters wide so \r + overwrite replaces it cleanly.
const statusLine = "Streaming live prices%s (ctrl+c to quit)"

func watchTable(ctx context.Context, updates <-chan *ws.CoinUpdate) error {
	warnf("Streaming live prices... (ctrl+c to quit)\n\n")

	state := make(map[string]*ws.CoinUpdate)
	prevPrices := make(map[string]float64)
	flashes := make(map[string]*priceFlash)
	var dotFrame int

	// Flash timer fires once to clear price color after flashDuration.
	var flashTimer *time.Timer
	var flashCh <-chan time.Time

	// Dot ticker animates the status line every 1s without redrawing the table.
	dotTicker := time.NewTicker(1 * time.Second)
	defer dotTicker.Stop()

	for {
		select {
		case update, ok := <-updates:
			if !ok {
				return nil
			}
			if prev, exists := prevPrices[update.CoinID]; exists {
				if update.Price > prev {
					flashes[update.CoinID] = &priceFlash{dir: 1, expires: time.Now().Add(flashDuration)}
				} else if update.Price < prev {
					flashes[update.CoinID] = &priceFlash{dir: -1, expires: time.Now().Add(flashDuration)}
				}
			}
			prevPrices[update.CoinID] = update.Price
			state[update.CoinID] = update
			renderWatchTable(state, flashes, dotFrame)

			// Schedule a single re-render to clear the flash color.
			if flashTimer != nil {
				flashTimer.Stop()
			}
			flashTimer = time.NewTimer(flashDuration)
			flashCh = flashTimer.C

		case <-flashCh:
			flashCh = nil
			renderWatchTable(state, flashes, dotFrame)

		case <-dotTicker.C:
			dotFrame = (dotFrame + 1) % len(statusDots)
			if len(state) > 0 && display.StdoutIsTerminal() {
				updateStatusLine(dotFrame)
			}

		case <-ctx.Done():
			if flashTimer != nil {
				flashTimer.Stop()
			}
			return nil
		}
	}
}

// updateStatusLine overwrites just the status line in place using ANSI save/restore
// cursor. No screen clear, no banner or table redraw.
func updateStatusLine(dotFrame int) {
	// Save cursor, move to the status line (banner rows + 1), overwrite, restore.
	row := display.BannerLines + 1
	fmt.Printf("\033[s\033[%d;1H\033[2K"+statusLine+"\033[u", row, statusDots[dotFrame])
}

// ANSI color constants for price flash.
const (
	colorFlashGreen = "\033[38;2;75;204;0m" // CoinGecko brand green (#4BCC00)
	colorFlashRed   = "\033[31m"
	colorReset      = "\033[0m"
)

// colorPercent formats a percentage with green/red color when colored is true.
// Uses the pre-computed stdout color flag instead of display.ColorPercent()
// which checks stderr.
func colorPercent(pct float64, colored bool) string {
	s := display.FormatPercent(pct)
	if !colored {
		return s
	}
	if pct > 0 {
		return colorFlashGreen + s + colorReset
	} else if pct < 0 {
		return colorFlashRed + s + colorReset
	}
	return s
}

// colorPrice formats the price, coloring it briefly when a flash is active.
// The colored flag should be pre-computed via display.StdoutColorEnabled() to
// avoid repeated syscalls in the render loop.
func colorPrice(price float64, flash *priceFlash, colored bool) string {
	s := display.FormatPrice(price)
	if flash == nil || !colored || !time.Now().Before(flash.expires) {
		return s
	}
	if flash.dir > 0 {
		return colorFlashGreen + s + colorReset
	}
	return colorFlashRed + s + colorReset
}

func renderWatchTable(state map[string]*ws.CoinUpdate, flashes map[string]*priceFlash, dotFrame int) {
	// Sort coin IDs for stable output.
	ids := make([]string, 0, len(state))
	for id := range state {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	// Cache color check once per render to avoid repeated syscalls.
	colored := display.StdoutColorEnabled()

	headers := []string{"Coin", "Price (USD)", "24h %", "Market Cap", "Volume (24h)", "Updated"}
	rows := make([][]string, 0, len(ids))
	for _, id := range ids {
		u := state[id]
		rows = append(rows, []string{
			display.SanitizeCell(id),
			colorPrice(u.Price, flashes[id], colored),
			colorPercent(u.Change24h, colored),
			display.FormatLargeNumber(u.MarketCap),
			display.FormatLargeNumber(u.Volume24h),
			formatTimestamp(u.UpdatedAt),
		})
	}

	// Clear previous output and re-render. Gate on stdout being an interactive
	// terminal, independent of color policy — screen control is not a color concern.
	// Banner and status are written to stdout (not stderr) so they're part of the
	// screen-clear cycle and don't spam stderr when it's redirected.
	if display.StdoutIsTerminal() {
		fmt.Print("\033[2J\033[H")
	}
	display.FprintBanner(os.Stdout)
	fmt.Printf(statusLine+"\n\n", statusDots[dotFrame])
	display.PrintTable(headers, rows)
}

// isBrokenPipe reports whether err is an EPIPE or ECONNRESET, indicating the
// downstream consumer closed the pipe (e.g. `cg watch | head -5`).
func isBrokenPipe(err error) bool {
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return errno == syscall.EPIPE || errno == syscall.ECONNRESET
	}
	return false
}

func formatTimestamp(ts int64) string {
	if ts == 0 {
		return "-"
	}
	return time.Unix(ts, 0).Format("15:04:05")
}

// matchSymbol picks the best coin ID from search results for a given symbol.
// It returns the exact case-insensitive match with the highest market_cap_rank,
// or "" if no match is found.
func matchSymbol(coins []coingecko.SearchCoin, symbol string) string {
	var best string
	var bestRank int
	for _, c := range coins {
		if !strings.EqualFold(c.Symbol, symbol) {
			continue
		}
		// market_cap_rank 0 means unranked — treat as worst.
		rank := c.MarketCapRank
		if rank == 0 {
			rank = 1<<31 - 1
		}
		if best == "" || rank < bestRank {
			best = c.ID
			bestRank = rank
		}
	}
	return best
}
