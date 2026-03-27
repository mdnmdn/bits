package cmd

import (
	"fmt"
	"os"

	legacycmd "github.com/mdnmdn/bits/internal/legacy/cmd"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "bits",
	Short: "bits CLI — cryptocurrency data at your fingertips",
}

func init() {
	RootCmd.PersistentFlags().StringP("output", "o", "table", "Output format (table, json, markdown, yaml)")
	RootCmd.PersistentFlags().StringP("provider", "p", "", "Data provider (coingecko, binance, bitget)")
	RootCmd.PersistentFlags().StringP("market", "m", "spot", "Market type (spot, futures, margin)")
	RootCmd.PersistentFlags().BoolP("lock", "l", false, "Disable provider fallback")
	RootCmd.AddCommand(legacycmd.LegacyCmd)
}

func Execute() {
	RootCmd.SilenceUsage = true
	RootCmd.SilenceErrors = true
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
