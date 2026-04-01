package cmd

import (
	"github.com/mdnmdn/bits/config"
	"github.com/mdnmdn/bits/internal/render"
	"github.com/mdnmdn/bits/model"
	"github.com/mdnmdn/bits/provider/registry"
	"github.com/mdnmdn/bits/resolve"
	"github.com/mdnmdn/bits/resolve/symbol"
	"github.com/spf13/cobra"
)

func loadConfig() (*config.Config, error) {
	cfg, _, err := config.Load()
	return cfg, err
}

func newResolver(cfg *config.Config) *resolve.Resolver {
	return resolve.New(cfg, registry.NewProvider, registry.AllProviderIDs)
}

func newSymbolEngine(cfg *config.Config) *symbol.SymbolEngine {
	return symbol.NewSymbolEngine(cfg)
}

func resolveOpts(cmd *cobra.Command) resolve.ResolutionOpts {
	provider, _ := cmd.Root().PersistentFlags().GetString("provider")
	market, _ := cmd.Root().PersistentFlags().GetString("market")
	allowFallback, _ := cmd.Root().PersistentFlags().GetBool("allow-fallback")
	return resolve.ResolutionOpts{
		Provider:   registry.ResolveProvider(provider),
		Market:     model.ParseMarketType(market),
		NoFallback: provider != "" && !allowFallback,
	}
}

func resolveFormat(cmd *cobra.Command) render.Format {
	f, _ := cmd.Root().PersistentFlags().GetString("output")
	return render.ParseFormat(f)
}
