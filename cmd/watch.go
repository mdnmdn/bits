package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/coingecko/coingecko-cli/internal/display"
	"github.com/coingecko/coingecko-cli/internal/ws"

	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Stream live coin prices via WebSocket (paid plans only)",
	Long:  "Connect to CoinGecko's real-time WebSocket API and stream live price updates. Requires a paid plan.",
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
		return fmt.Errorf("provide --ids or --symbols")
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	// Resolve coin IDs.
	var coinIDs []string
	if idsStr != "" {
		coinIDs = append(coinIDs, splitTrim(idsStr)...)
	}

	if symbolsStr != "" {
		symbols := splitTrim(symbolsStr)
		client := newAPIClient(cfg)
		prices, err := client.SimplePriceBySymbols(cmd.Context(), symbols, "usd")
		if err != nil {
			return fmt.Errorf("resolving symbols: %w", err)
		}
		for id := range prices {
			coinIDs = append(coinIDs, id)
		}
		if len(coinIDs) == 0 {
			return fmt.Errorf("no coins found for symbols: %s", symbolsStr)
		}
	}

	if isDryRun(cmd) {
		return printDryRunWS(cfg, coinIDs)
	}

	// Set up context with signal handling.
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
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
	// Save cursor position, move to the status line (known position after
	// banner: \n=row1, text=row2, \n\n=row3, status=row4), overwrite it,
	// then restore cursor so the terminal stays visually stable.
	fmt.Printf("\033[s\033[4;1H\033[2K"+statusLine+"\033[u", statusDots[dotFrame])
}

// ANSI color constants for price flash.
const (
	colorBrandGreen = "\033[38;2;140;195;81m" // CoinGecko brand green (#8CC351)
	colorFlashRed   = "\033[31m"
	colorReset      = "\033[0m"
)

// colorPrice formats the price, coloring it briefly when a flash is active.
// The colored flag should be pre-computed via display.StdoutColorEnabled() to
// avoid repeated syscalls in the render loop.
func colorPrice(price float64, flash *priceFlash, colored bool) string {
	s := display.FormatPrice(price)
	if flash == nil || !colored || !time.Now().Before(flash.expires) {
		return s
	}
	if flash.dir > 0 {
		return colorBrandGreen + s + colorReset
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
			display.ColorPercent(u.Change24h),
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
