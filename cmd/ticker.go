package cmd

import (
	"fmt"

	"github.com/mdnmdn/bits/internal/display"
	"github.com/mdnmdn/bits/internal/provider"

	"github.com/spf13/cobra"
)

var tickerCmd = &cobra.Command{
	Use:   "ticker [symbol]",
	Short: "Show 24h ticker stats for a trading pair",
	Long:  "Fetch 24-hour ticker statistics from exchange providers (Binance, Bitget).",
	Example: `  bits ticker BTCUSDT -p binance
  bits ticker ETHUSDT -p bitget
  bits ticker BTCUSDT -p binance -o json`,
	Args: cobra.ExactArgs(1),
	RunE: runTicker,
}

func init() {
	rootCmd.AddCommand(tickerCmd)
}

func runTicker(cmd *cobra.Command, args []string) error {
	symbol := args[0]
	jsonOut := outputJSON(cmd)

	if !jsonOut {
		display.PrintBanner()
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	client := newAPIClient(cfg)

	tp, ok := client.(provider.TickerProvider)
	if !ok {
		return fmt.Errorf("%s provider does not support ticker data — use -p binance or -p bitget", client.ID())
	}

	ticker, err := tp.Ticker24h(cmd.Context(), symbol)
	if err != nil {
		return err
	}

	if jsonOut {
		return printJSONRaw(ticker)
	}

	headers := []string{"Metric", "Value"}
	rows := [][]string{
		{"Symbol", ticker.Symbol},
		{"Last Price", display.FormatPrice(ticker.LastPrice)},
		{"24h Change", display.ColorPercent(ticker.PriceChangePercent)},
		{"24h High", display.FormatPrice(ticker.HighPrice)},
		{"24h Low", display.FormatPrice(ticker.LowPrice)},
		{"Volume", display.FormatLargeNumber(ticker.Volume)},
		{"Quote Volume", display.FormatLargeNumber(ticker.QuoteVolume)},
	}
	display.PrintTable(headers, rows)
	return nil
}
