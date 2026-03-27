package cmd

import (
	"os"

	"github.com/mdnmdn/bits/internal/capability"
	renderjson "github.com/mdnmdn/bits/internal/render/json"
	rendertable "github.com/mdnmdn/bits/internal/render/table"
	"github.com/mdnmdn/bits/internal/resolve"
	"github.com/spf13/cobra"

	"github.com/mdnmdn/bits/internal/provider"
)

var bookCmd = &cobra.Command{
	Use:   "book <symbol>",
	Short: "Order book snapshot",
	Args:  cobra.ExactArgs(1),
	RunE:  runBook,
}

func init() {
	bookCmd.Flags().Int("depth", 20, "Order book depth")
	RootCmd.AddCommand(bookCmd)
}

func runBook(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	opts := resolveOpts(cmd)
	format := resolveFormat(cmd)
	depth, _ := cmd.Flags().GetInt("depth")
	resolver := newResolver(cfg)

	p, market, fallback, rerr := resolver.Resolve(cmd.Context(), capability.FeatureOrderBook, resolve.ResolutionOpts{
		Provider: opts.Provider,
		Market:   opts.Market,
		Lock:     opts.Lock,
	})
	if rerr != nil {
		return rerr
	}

	obp, rerr := resolve.Require[provider.OrderBookProvider](p, "order-book")
	if rerr != nil {
		return rerr
	}

	res, err := obp.OrderBook(cmd.Context(), args[0], market, depth)
	if err != nil {
		return err
	}

	if fallback {
		res.Fallback = true
		res.RequestedProvider = opts.Provider
		res.RequestedMarket = opts.Market
	}
	res.Market = market

	switch format {
	case "json":
		return renderjson.Render(os.Stdout, res)
	default:
		return rendertable.RenderOrderBook(os.Stdout, res)
	}
}
