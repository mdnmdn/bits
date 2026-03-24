package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/coingecko/coingecko-cli/internal/provider/coingecko"
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

	if !jsonOut {
		display.PrintBanner()
	}

	if idsStr == "" && symbolsStr == "" {
		return fmt.Errorf("provide --ids or --symbols")
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	// Short-circuit before any API calls in dry-run mode.
	if isDryRun(cmd) {
		params := map[string]string{
			"vs_currencies":       vs,
			"include_24hr_change": "true",
		}
		if idsStr != "" {
			params["ids"] = idsStr
		}
		if symbolsStr != "" {
			params["symbols"] = symbolsStr
		}
		return printDryRun(cfg, "price", "/simple/price", params, nil)
	}

	client := newAPIClient(cfg)
	ctx := cmd.Context()

	// Fetch prices by IDs and/or symbols directly — no /search resolution needed.
	// The API's symbols lookup returns the top-ranked match by market cap.
	var prices coingecko.PriceResponse
	var requestedIDs []string // track user-requested IDs for missing-coin warnings (symbols can't be checked — response keys are coin IDs)

	if idsStr != "" {
		ids := splitTrim(idsStr)
		requestedIDs = append(requestedIDs, ids...)
		p, err := client.SimplePrice(ctx, ids, vs)
		if err != nil {
			return err
		}
		prices = p
	}

	if symbolsStr != "" {
		symbols := splitTrim(symbolsStr)
		p, err := client.SimplePriceBySymbols(ctx, symbols, vs)
		if err != nil {
			return err
		}
		if prices == nil {
			prices = p
		} else {
			for k, v := range p {
				prices[k] = v
			}
		}
	}

	if len(prices) == 0 {
		return fmt.Errorf("no valid coins found")
	}

	if jsonOut {
		return printJSONRaw(prices)
	}

	// Warn about requested IDs that returned no data.
	// Only check --ids values; --symbols can't be checked because response keys are coin IDs (e.g. "bitcoin"), not symbols (e.g. "btc").
	responseKeys := make(map[string]bool, len(prices))
	for k := range prices {
		responseKeys[strings.ToLower(k)] = true
	}
	for _, r := range requestedIDs {
		if !responseKeys[strings.ToLower(r)] {
			warnf("Warning: no data returned for %q\n", r)
		}
	}

	// Sort response keys for deterministic table output.
	keys := make([]string, 0, len(prices))
	for k := range prices {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	headers := []string{"Coin", "Price", "24h Change"}
	var rows [][]string
	for _, id := range keys {
		data := prices[id]
		rows = append(rows, []string{
			display.SanitizeCell(id),
			display.FormatPrice(data[vs], vs),
			display.ColorPercent(data[vs+"_24h_change"]),
		})
	}

	display.PrintTable(headers, rows)
	return nil
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
