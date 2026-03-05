package cmd

import (
	"os"

	"github.com/coingecko/coingecko-cli/internal/display"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:     "cg",
	Short:   "CoinGecko CLI — cryptocurrency data at your fingertips",
	Long:    "A command-line tool for accessing CoinGecko cryptocurrency market data.",
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		display.PrintLogo()
		display.PrintWelcomeBox()
	},
}

func init() {
	rootCmd.PersistentFlags().StringP("output", "o", "table", "Output format (table, json)")
}

func Execute() {
	rootCmd.SilenceUsage = true
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func outputJSON(cmd *cobra.Command) bool {
	o, _ := cmd.Flags().GetString("output")
	return o == "json"
}
