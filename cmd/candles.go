package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/mdnmdn/bits/capability"
	rendertable "github.com/mdnmdn/bits/internal/render/table"
	"github.com/mdnmdn/bits/model"
	"github.com/mdnmdn/bits/resolve"
	"github.com/spf13/cobra"

	"github.com/mdnmdn/bits/provider"
)

var candlesCmd = &cobra.Command{
	Use:   "candles <symbol>",
	Short: "OHLCV history",
	Args:  cobra.ExactArgs(1),
	RunE:  runCandles,
}

func init() {
	candlesCmd.Flags().String("interval", "1h", "Candle interval (e.g. 1m, 5m, 1h, 1d)")
	candlesCmd.Flags().String("from", "", "Start time (RFC3339 or YYYY-MM-DD)")
	candlesCmd.Flags().String("to", "", "End time (RFC3339 or YYYY-MM-DD)")
	candlesCmd.Flags().Int("limit", 0, "Maximum number of candles (0 = provider default)")
	RootCmd.AddCommand(candlesCmd)
}

func runCandles(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	opts := resolveOpts(cmd)
	format := resolveFormat(cmd)
	interval, _ := cmd.Flags().GetString("interval")
	fromStr, _ := cmd.Flags().GetString("from")
	toStr, _ := cmd.Flags().GetString("to")
	limit, _ := cmd.Flags().GetInt("limit")
	resolver := newResolver(cfg)

	candleOpts := model.CandleOpts{}
	if fromStr != "" {
		t, terr := parseTime(fromStr)
		if terr != nil {
			return terr
		}
		candleOpts.From = &t
	}
	if toStr != "" {
		t, terr := parseTime(toStr)
		if terr != nil {
			return terr
		}
		candleOpts.To = &t
	}
	if limit > 0 {
		candleOpts.Limit = &limit
	}

	p, market, fallback, rerr := resolver.Resolve(cmd.Context(), capability.FeatureCandles, resolve.ResolutionOpts{
		Provider:   opts.Provider,
		Market:     opts.Market,
		NoFallback: opts.NoFallback,
	})
	if rerr != nil {
		return rerr
	}

	cp, rerr := resolve.Require[provider.CandleProvider](p, "candles")
	if rerr != nil {
		return rerr
	}

	symEngine := newSymbolEngine(cfg)
	providerID := p.ID()
	symbol := args[0]
	resolved, err := symEngine.Resolve(cmd.Context(), providerID, symbol, market)
	if err == nil && resolved != "" {
		symbol = resolved
	}

	res, err := cp.Candles(cmd.Context(), symbol, market, interval, candleOpts)
	if err != nil {
		return err
	}

	if fallback {
		res.Fallback = true
		res.RequestedProvider = opts.Provider
		res.RequestedMarket = opts.Market
	}
	res.Market = market

	if ok, err := renderGeneric(os.Stdout, format, res); ok || err != nil {
		return err
	}
	return rendertable.RenderCandles(os.Stdout, res)
}

func parseTime(s string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognised time format %q (use RFC3339 or YYYY-MM-DD)", s)
}
