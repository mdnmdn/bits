package command

import (
	"github.com/spf13/cobra"
)

var Root = &cobra.Command{
	Use:   "bits",
	Short: "bits CLI — cryptocurrency data at your fingertips",
}

func init() {
	Root.PersistentFlags().StringP("log", "l", "info", "Log level (debug, info, warn, error)")
	Root.PersistentFlags().StringP("output", "o", "table", "Output format (table, json, markdown, yaml, toon)")
	Root.PersistentFlags().StringP("provider", "p", "", "Data provider (coingecko, binance, bitget, whitebit, cryptocom, mexc)")
	Root.PersistentFlags().StringP("market", "m", "spot", "Market type (spot, futures/future, margin)")
}

func AddCommand(cmd *cobra.Command) {
	Root.AddCommand(cmd)
}

func AddCommands(cmds ...*cobra.Command) {
	for _, cmd := range cmds {
		Root.AddCommand(cmd)
	}
}
