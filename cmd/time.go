package cmd

import (
	"os"

	"github.com/mdnmdn/bits/pkg/capability"
	"github.com/mdnmdn/bits/internal/process"
	rendertable "github.com/mdnmdn/bits/internal/render/table"
	"github.com/mdnmdn/bits/pkg/resolve"
	"github.com/spf13/cobra"

	"github.com/mdnmdn/bits/pkg/provider"
)

var timeCmd = &cobra.Command{
	Use:   "time",
	Short: "Show exchange server time",
	RunE:  runTime,
}

func init() {
	RootCmd.AddCommand(timeCmd)
}

func runTime(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	opts := resolveOpts(cmd)
	format := resolveFormat(cmd)
	resolver := newResolver(cfg)

	p, market, fallback, rerr := resolver.Resolve(cmd.Context(), capability.FeatureServerTime, resolve.ResolutionOpts{
		Provider:   opts.Provider,
		Market:     opts.Market,
		NoFallback: opts.NoFallback,
	})
	if rerr != nil {
		return rerr
	}

	ep, rerr := resolve.Require[provider.ExchangeProvider](p, "server-time")
	if rerr != nil {
		return rerr
	}

	res, err := ep.ServerTime(cmd.Context())
	if err != nil {
		return err
	}

	if fallback {
		res.Fallback = true
		res.RequestedProvider = opts.Provider
		res.RequestedMarket = opts.Market
	}
	res.Market = market

	res = process.Apply(res, process.TimeEnricher)

	if ok, err := renderGeneric(os.Stdout, format, res); ok || err != nil {
		return err
	}
	return rendertable.RenderServerTime(os.Stdout, res)
}
