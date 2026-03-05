package cmd

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/coingecko/coingecko-cli/internal/api"
	"github.com/coingecko/coingecko-cli/internal/config"
	"github.com/coingecko/coingecko-cli/internal/display"

	"github.com/spf13/cobra"
)

var priceCmd = &cobra.Command{
	Use:   "price",
	Short: "Get current price for coins",
	Long:  "Fetch current prices by coin IDs or symbols. Use --ids for CoinGecko IDs (e.g. bitcoin) or --symbols for ticker symbols (e.g. btc).",
	Example: `  cg price --ids bitcoin,ethereum
  cg price --symbols btc,eth --vs eur
  cg price --ids solana -o json`,
	RunE: runPrice,
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
	ctx := cmd.Context()

	if !jsonOut {
		display.PrintBanner()
	}

	if idsStr == "" && symbolsStr == "" {
		return fmt.Errorf("provide --ids or --symbols")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Short-circuit before any API calls in dry-run mode.
	if isDryRun(cmd) {
		raw := idsStr
		if symbolsStr != "" {
			if raw != "" {
				raw += ","
			}
			raw += symbolsStr
		}
		return printDryRun(cfg, "price", "/simple/price", map[string]string{
			"ids":                 raw,
			"vs_currencies":       vs,
			"include_24hr_change": "true",
		}, nil)
	}

	client := api.NewClient(cfg)

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
		return printJSONRaw(prices)
	}

	headers := []string{"Coin", "Price", "24h Change"}
	var rows [][]string
	for _, id := range ids {
		data, ok := prices[id]
		if !ok {
			warnf("Warning: no data returned for %q\n", id)
			continue
		}
		price := data[vs]
		change := data[vs+"_24h_change"]
		rows = append(rows, []string{
			display.SanitizeCell(id),
			display.FormatPrice(price, vs),
			display.ColorPercent(change),
		})
	}

	display.PrintTable(headers, rows)
	return nil
}

func resolveSymbols(ctx context.Context, client *api.Client, symbols []string) ([]string, error) {
	type result struct {
		index int
		id    string
		sym   string
		err   error
	}

	results := make([]result, len(symbols))
	var wg sync.WaitGroup
	for i, sym := range symbols {
		wg.Add(1)
		go func(idx int, sym string) {
			defer wg.Done()
			sym = strings.ToLower(sym)
			resp, err := client.Search(ctx, sym)
			if err != nil {
				results[idx] = result{index: idx, sym: sym, err: err}
				return
			}
			var best *api.SearchCoin
			for j, c := range resp.Coins {
				if strings.EqualFold(c.Symbol, sym) {
					if best == nil || (c.MarketCapRank > 0 && (best.MarketCapRank == 0 || c.MarketCapRank < best.MarketCapRank)) {
						best = &resp.Coins[j]
					}
				}
			}
			if best != nil {
				results[idx] = result{index: idx, id: best.ID, sym: sym}
			} else {
				results[idx] = result{index: idx, sym: sym}
			}
		}(i, sym)
	}
	wg.Wait()

	var ids []string
	for _, r := range results {
		if r.err != nil {
			return nil, fmt.Errorf("searching for %q: %w", r.sym, r.err)
		}
		if r.id == "" {
			warnf("Warning: no exact match for symbol %q, skipping\n", r.sym)
			continue
		}
		ids = append(ids, r.id)
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
