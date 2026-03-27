package cmd

import "github.com/spf13/cobra"

var CoingeckoCmd = &cobra.Command{
	Use:   "coingecko",
	Short: "CoinGecko-only commands",
	Long:  "Commands that are only available on CoinGecko provider.",
}

func init() {
	CoingeckoCmd.AddCommand(marketsCmd, searchCmd, trendingCmd, historyCmd, topGainersLosersCmd)
	RootCmd.AddCommand(CoingeckoCmd)
}
