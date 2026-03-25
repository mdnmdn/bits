package cmd

import (
	"fmt"
	"os"

	"github.com/mdnmdn/bits/internal/display"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:     "bits",
	Short:   "bits CLI — cryptocurrency data at your fingertips",
	Long:    "A command-line tool for accessing multi-provider cryptocurrency market data.",
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		display.PrintLogo()
		display.PrintWelcomeBox()
	},
}

func init() {
	rootCmd.PersistentFlags().StringP("output", "o", "table", "Output format (table, json)")
	rootCmd.PersistentFlags().StringP("provider", "p", "", "Data provider (coingecko, binance, bitget)")
	rootCmd.PersistentFlags().StringP("market-type", "m", "", "Market type (spot, margin, future)")
}

func Execute() {
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	if err := rootCmd.Execute(); err != nil {
		// Emit structured JSON error to stderr when -o json, otherwise plain text.
		cmd, _, _ := rootCmd.Find(os.Args[1:])
		if cmd != nil && outputJSON(cmd) {
			_ = formatError(cmd, err)
		} else {
			fmt.Fprintln(os.Stderr, "Error:", err)
		}
		os.Exit(1)
	}
}

func outputJSON(cmd *cobra.Command) bool {
	o, _ := cmd.Flags().GetString("output")
	return o == "json"
}
