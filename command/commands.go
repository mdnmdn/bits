package command

import (
	"github.com/mdnmdn/bits/internal/logger"
	"github.com/spf13/cobra"
)

var Root = &cobra.Command{
	Use:   "bits",
	Short: "bits CLI — cryptocurrency data at your fingertips",
}

var logLevel string

func init() {
	Root.PersistentFlags().StringVarP(&logLevel, "log", "l", "info", "Log level (debug, info, warn, error)")
	Root.PersistentFlags().StringP("output", "o", "table", "Output format (table, json, markdown, yaml, toon)")
	Root.PersistentFlags().StringP("provider", "p", "", "Data provider (coingecko, binance, bitget, whitebit, cryptocom, mexc)")
	Root.PersistentFlags().StringP("market", "m", "spot", "Market type (spot, futures/future, margin)")

	Root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		logger.SetLevel(logLevel)
		return nil
	}
}

func AddCommand(cmd *cobra.Command) {
	Root.AddCommand(cmd)
}

func AddCommands(cmds ...*cobra.Command) {
	for _, cmd := range cmds {
		Root.AddCommand(cmd)
	}
}
