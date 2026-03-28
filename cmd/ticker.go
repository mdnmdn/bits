package cmd

import (
	"context"
	"os"

	"github.com/mdnmdn/bits/internal/capability"
	"github.com/mdnmdn/bits/internal/model"
	rendertable "github.com/mdnmdn/bits/internal/render/table"
	"github.com/mdnmdn/bits/internal/resolve"
	"github.com/spf13/cobra"

	"github.com/mdnmdn/bits/internal/provider"
)

var tickerCmd = &cobra.Command{
	Use:   "ticker <symbol>...",
	Short: "24h rolling statistics",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runTicker,
}

func init() {
	RootCmd.AddCommand(tickerCmd)
}

func runTicker(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	opts := resolveOpts(cmd)
	format := resolveFormat(cmd)
	resolver := newResolver(cfg)

	p, market, fallback, rerr := resolver.Resolve(cmd.Context(), capability.FeatureTicker24h, resolve.ResolutionOpts{
		Provider:   opts.Provider,
		Market:     opts.Market,
		NoFallback: opts.NoFallback,
	})
	if rerr != nil {
		return rerr
	}

	tp, rerr := resolve.Require[provider.TickerProvider](p, "ticker")
	if rerr != nil {
		return rerr
	}

	res := resolve.FanOut(cmd.Context(), args, func(ctx context.Context, symbol string) (model.Response[model.Ticker24h], error) {
		return tp.Ticker24h(ctx, symbol, market)
	})

	if fallback {
		res.Fallback = true
		res.RequestedProvider = opts.Provider
		res.RequestedMarket = opts.Market
	}
	res.Market = market

	if ok, err := renderGeneric(os.Stdout, format, res); ok || err != nil {
		return err
	}
	return rendertable.RenderTickers(os.Stdout, res)
}
