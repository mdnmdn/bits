package cmd

import (
	"fmt"

	"github.com/coingecko/coingecko-cli/internal/display"
	"github.com/coingecko/coingecko-cli/internal/provider"

	"github.com/spf13/cobra"
)

var orderbookCmd = &cobra.Command{
	Use:   "orderbook [symbol]",
	Short: "Show order book for a trading pair",
	Long:  "Fetch order book depth from exchange providers (Binance).",
	Example: `  bits orderbook BTCUSDT -p binance
  bits orderbook ETHUSDT -p binance --limit 10
  bits orderbook BTCUSDT -p binance -o json`,
	Args: cobra.ExactArgs(1),
	RunE: runOrderbook,
}

func init() {
	orderbookCmd.Flags().Int("limit", 20, "Number of bid/ask levels to show")
	rootCmd.AddCommand(orderbookCmd)
}

func runOrderbook(cmd *cobra.Command, args []string) error {
	symbol := args[0]
	limit, _ := cmd.Flags().GetInt("limit")
	jsonOut := outputJSON(cmd)

	if !jsonOut {
		display.PrintBanner()
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	client := newAPIClient(cfg)

	obp, ok := client.(provider.OrderBookProvider)
	if !ok {
		return fmt.Errorf("%s provider does not support order book data — use -p binance", client.ID())
	}

	ob, err := obp.OrderBook(cmd.Context(), symbol, limit)
	if err != nil {
		return err
	}

	if jsonOut {
		return printJSONRaw(ob)
	}

	// Show bids
	fmt.Printf("Order Book: %s\n\n", symbol)
	fmt.Println("Bids (Buy)")
	bidHeaders := []string{"Price", "Quantity"}
	bidRows := make([][]string, 0, len(ob.Bids))
	showLimit := limit
	if showLimit > len(ob.Bids) {
		showLimit = len(ob.Bids)
	}
	for i := 0; i < showLimit; i++ {
		bidRows = append(bidRows, []string{
			display.FormatPrice(ob.Bids[i].Price),
			fmt.Sprintf("%.8f", ob.Bids[i].Quantity),
		})
	}
	display.PrintTable(bidHeaders, bidRows)

	// Show asks
	fmt.Println()
	fmt.Println("Asks (Sell)")
	askHeaders := []string{"Price", "Quantity"}
	askRows := make([][]string, 0, len(ob.Asks))
	showLimit = limit
	if showLimit > len(ob.Asks) {
		showLimit = len(ob.Asks)
	}
	for i := 0; i < showLimit; i++ {
		askRows = append(askRows, []string{
			display.FormatPrice(ob.Asks[i].Price),
			fmt.Sprintf("%.8f", ob.Asks[i].Quantity),
		})
	}
	display.PrintTable(askHeaders, askRows)
	return nil
}
