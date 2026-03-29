package cmd

import (
	"os"

	"github.com/mdnmdn/bits/pkg/capability"
	"github.com/mdnmdn/bits/pkg/model"
	rendertable "github.com/mdnmdn/bits/internal/render/table"
	"github.com/mdnmdn/bits/pkg/resolve"
	"github.com/spf13/cobra"

	"github.com/mdnmdn/bits/pkg/provider"
)

var marketsCmd = &cobra.Command{
	Use:   "markets",
	Short: "Ranked coin listing (aggregators only)",
	RunE:  runMarkets,
}

func init() {
	marketsCmd.Flags().String("currency", "usd", "Quote currency")
	marketsCmd.Flags().Int("page", 1, "Page number")
	marketsCmd.Flags().Int("per-page", 100, "Results per page")
	marketsCmd.Flags().String("order", "market_cap_desc", "Sort order")
	marketsCmd.Flags().String("category", "", "Coin category filter")
	RootCmd.AddCommand(marketsCmd)
}

func runMarkets(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	opts := resolveOpts(cmd)
	format := resolveFormat(cmd)
	currency, _ := cmd.Flags().GetString("currency")
	page, _ := cmd.Flags().GetInt("page")
	perPage, _ := cmd.Flags().GetInt("per-page")
	order, _ := cmd.Flags().GetString("order")
	category, _ := cmd.Flags().GetString("category")
	resolver := newResolver(cfg)

	p, market, fallback, rerr := resolver.Resolve(cmd.Context(), capability.FeatureMarketsList, resolve.ResolutionOpts{
		Provider:   opts.Provider,
		Market:     opts.Market,
		NoFallback: opts.NoFallback,
	})
	if rerr != nil {
		return rerr
	}

	ap, rerr := resolve.Require[provider.AggregatorProvider](p, "markets")
	if rerr != nil {
		return rerr
	}

	res, err := ap.CoinMarkets(cmd.Context(), model.MarketOpts{
		Currency: currency,
		PerPage:  perPage,
		Page:     page,
		Order:    order,
		Category: category,
	})
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
	return rendertable.RenderMarkets(os.Stdout, res)
}
