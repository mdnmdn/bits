package cmd

import (
	"context"
	"fmt"
	"strings"

	"coingecko-cli/internal/api"
	"coingecko-cli/internal/config"
	"coingecko-cli/internal/display"

	"github.com/spf13/cobra"
)

var priceCmd = &cobra.Command{
	Use:   "price",
	Short: "Get current price for coins",
	Long:  "Fetch current prices by coin IDs or symbols. Use --ids for CoinGecko IDs (e.g. bitcoin) or --symbols for ticker symbols (e.g. btc).",
	RunE:  runPrice,
}

func init() {
	priceCmd.Flags().String("ids", "", "Comma-separated coin IDs (e.g. bitcoin,ethereum)")
	priceCmd.Flags().String("symbols", "", "Comma-separated symbols (e.g. btc,eth)")
	priceCmd.Flags().String("vs", "usd", "Target currency")
	rootCmd.AddCommand(priceCmd)
}

func runPrice(cmd *cobra.Command, args []string) error {
	idsStr, _ := cmd.Flags().GetString("ids")
	symbolsStr, _ := cmd.Flags().GetString("symbols")
	vs, _ := cmd.Flags().GetString("vs")
	jsonOut := outputJSON(cmd)

	if idsStr == "" && symbolsStr == "" {
		return fmt.Errorf("provide --ids or --symbols")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	client := api.NewClient(cfg)
	ctx := context.Background()

	var ids []string
	if idsStr != "" {
		ids = splitTrim(idsStr)
	}

	if symbolsStr != "" {
		symbols := splitTrim(symbolsStr)
		resolved, err := resolveSymbols(ctx, client, symbols)
		if err != nil {
			return err
		}
		ids = append(ids, resolved...)
	}

	if len(ids) == 0 {
		return fmt.Errorf("no valid coins found")
	}

	prices, err := client.SimplePrice(ctx, ids, vs)
	if err != nil {
		return err
	}

	if jsonOut {
		printJSONRaw(prices)
		return nil
	}

	headers := []string{"Coin", "Price", "24h Change"}
	var rows [][]string
	for _, id := range ids {
		data, ok := prices[id]
		if !ok {
			continue
		}
		price, _ := toFloat(data[vs])
		changeKey := vs + "_24h_change"
		change, _ := toFloat(data[changeKey])
		rows = append(rows, []string{
			id,
			display.FormatPrice(price),
			display.ColorPercent(change),
		})
	}

	display.PrintTable(headers, rows)
	return nil
}

func resolveSymbols(ctx context.Context, client *api.Client, symbols []string) ([]string, error) {
	var ids []string
	for _, sym := range symbols {
		sym = strings.ToLower(sym)
		resp, err := client.Search(ctx, sym)
		if err != nil {
			return nil, fmt.Errorf("searching for %q: %w", sym, err)
		}
		var best *api.SearchCoin
		for i, c := range resp.Coins {
			if strings.EqualFold(c.Symbol, sym) {
				if best == nil || (c.MarketCapRank > 0 && (best.MarketCapRank == 0 || c.MarketCapRank < best.MarketCapRank)) {
					best = &resp.Coins[i]
				}
			}
		}
		if best == nil {
			warnf("Warning: no exact match for symbol %q, skipping\n", sym)
			continue
		}
		ids = append(ids, best.ID)
	}
	return ids, nil
}

func splitTrim(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func toFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	default:
		return 0, false
	}
}
