package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "bits",
	Short: "bits CLI — cryptocurrency data at your fingertips",
}

func init() {
	RootCmd.PersistentFlags().StringP("output", "o", "table", "Output format (table, json, markdown, yaml, toon)")
	RootCmd.PersistentFlags().StringP("provider", "p", "", "Data provider (coingecko, binance, bitget)")
	RootCmd.PersistentFlags().StringP("market", "m", "spot", "Market type (spot, futures, margin)")
	RootCmd.PersistentFlags().BoolP("allow-fallback", "f", false, "Allow provider fallback even when --provider is set")
}

func Execute() {
	RootCmd.SilenceUsage = true
	RootCmd.SilenceErrors = true
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
