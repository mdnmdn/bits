package cmd

import (
	"github.com/mdnmdn/bits/internal/config"
	"github.com/mdnmdn/bits/internal/model"
	"github.com/mdnmdn/bits/internal/registry"
	"github.com/mdnmdn/bits/internal/render"
	"github.com/mdnmdn/bits/internal/resolve"
	"github.com/spf13/cobra"
)

func loadConfig() (*config.Config, error) {
	cfg, _, err := config.Load()
	return cfg, err
}

func newResolver(cfg *config.Config) *resolve.Resolver {
	return resolve.New(cfg, registry.NewProvider, registry.AllProviderIDs)
}

func resolveOpts(cmd *cobra.Command) resolve.ResolutionOpts {
	provider, _ := cmd.Root().PersistentFlags().GetString("provider")
	market, _ := cmd.Root().PersistentFlags().GetString("market")
	allowFallback, _ := cmd.Root().PersistentFlags().GetBool("allow-fallback")
	return resolve.ResolutionOpts{
		Provider:   provider,
		Market:     model.MarketType(market),
		NoFallback: provider != "" && !allowFallback,
	}
}

func resolveFormat(cmd *cobra.Command) render.Format {
	f, _ := cmd.Root().PersistentFlags().GetString("output")
	return render.ParseFormat(f)
}
