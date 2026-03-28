package cmd

import (
	"os"

	"github.com/mdnmdn/bits/internal/capability"
	rendertable "github.com/mdnmdn/bits/internal/render/table"
	"github.com/mdnmdn/bits/internal/resolve"
	"github.com/spf13/cobra"

	"github.com/mdnmdn/bits/internal/provider"
)

var priceCmd = &cobra.Command{
	Use:   "price <id|symbol>...",
	Short: "Current price(s)",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runPrice,
}

func init() {
	priceCmd.Flags().String("currency", "usd", "Quote currency")
	RootCmd.AddCommand(priceCmd)
}

func runPrice(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	opts := resolveOpts(cmd)
	format := resolveFormat(cmd)
	currency, _ := cmd.Flags().GetString("currency")
	resolver := newResolver(cfg)

	p, market, fallback, rerr := resolver.Resolve(cmd.Context(), capability.FeaturePrice, resolve.ResolutionOpts{
		Provider: opts.Provider,
		Market:   opts.Market,
		Lock:     opts.Lock,
	})
	if rerr != nil {
		return rerr
	}

	pp, rerr := resolve.Require[provider.PriceProvider](p, "price")
	if rerr != nil {
		return rerr
	}

	res, err := pp.Price(cmd.Context(), args, currency)
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
	return rendertable.RenderPrices(os.Stdout, res)
}
