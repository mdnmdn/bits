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

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Exchange symbol catalogue",
	RunE:  runInfo,
}

func init() {
	infoCmd.Flags().String("symbol", "", "Filter by symbol")
	RootCmd.AddCommand(infoCmd)
}

func runInfo(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	opts := resolveOpts(cmd)
	format := resolveFormat(cmd)
	symbolFilter, _ := cmd.Flags().GetString("symbol")
	resolver := newResolver(cfg)

	p, market, fallback, rerr := resolver.Resolve(cmd.Context(), capability.FeatureExchangeInfo, resolve.ResolutionOpts{
		Provider: opts.Provider,
		Market:   opts.Market,
		Lock:     opts.Lock,
	})
	if rerr != nil {
		return rerr
	}

	ep, rerr := resolve.Require[provider.ExchangeProvider](p, "exchange-info")
	if rerr != nil {
		return rerr
	}

	res, err := ep.ExchangeInfo(cmd.Context(), market)
	if err != nil {
		return err
	}

	if fallback {
		res.Fallback = true
		res.RequestedProvider = opts.Provider
		res.RequestedMarket = opts.Market
	}
	res.Market = market

	if symbolFilter != "" {
		filtered := res.Data.Symbols[:0]
		for _, s := range res.Data.Symbols {
			if s.Symbol == symbolFilter {
				filtered = append(filtered, s)
			}
		}
		res.Data.Symbols = filtered
	}

	switch format {
	case "json":
		return renderjson.Render(os.Stdout, res)
	default:
		return rendertable.RenderExchangeInfo(os.Stdout, res)
	}
}
