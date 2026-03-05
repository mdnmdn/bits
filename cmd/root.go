package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "cg",
	Short:   "CoinGecko CLI — cryptocurrency data at your fingertips",
	Long:    "A command-line tool for accessing CoinGecko cryptocurrency market data.",
	Version: version,
}

func Execute() {
	rootCmd.SilenceUsage = true
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
