package cmd

import (
	"os"

	rendertable "github.com/mdnmdn/bits/internal/render/table"
	"github.com/mdnmdn/bits/pkg/capability"
	"github.com/mdnmdn/bits/pkg/resolve"
	"github.com/spf13/cobra"

	"github.com/mdnmdn/bits/pkg/provider"
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
		Provider:   opts.Provider,
		Market:     opts.Market,
		NoFallback: opts.NoFallback,
	})
	if rerr != nil {
		return rerr
	}

	obp, rerr := resolve.Require[provider.OrderBookProvider](p, "order-book")
	if rerr != nil {
		return rerr
	}

	symResolver := newSymbolResolver(p)
	symbol := args[0]
	resolved, err := symResolver.Resolve(cmd.Context(), symbol, market)
	if err == nil && resolved != "" {
		symbol = resolved
	}

	res, err := obp.OrderBook(cmd.Context(), symbol, market, depth)
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
	return rendertable.RenderOrderBook(os.Stdout, res)
}
