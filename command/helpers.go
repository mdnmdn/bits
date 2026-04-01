package command

import (
	"io"

	"github.com/mdnmdn/bits/config"
	"github.com/mdnmdn/bits/model"
	"github.com/mdnmdn/bits/provider/registry"
	"github.com/mdnmdn/bits/render"
	renderjson "github.com/mdnmdn/bits/render/json"
	rendermd "github.com/mdnmdn/bits/render/markdown"
	rendertoon "github.com/mdnmdn/bits/render/toon"
	renderyaml "github.com/mdnmdn/bits/render/yaml"
	"github.com/spf13/cobra"
)

func LoadConfig() (*config.Config, error) {
	cfg, _, err := config.Load()
	return cfg, err
}

func ResolveOptions(cmd *cobra.Command) (string, model.MarketType, render.Format, error) {
	provider, _ := cmd.Root().PersistentFlags().GetString("provider")
	market, _ := cmd.Root().PersistentFlags().GetString("market")

	providerID := registry.ResolveProvider(provider)
	mkt := model.ParseMarketType(market)

	formatStr, _ := cmd.Root().PersistentFlags().GetString("output")
	format := render.ParseFormat(formatStr)

	return providerID, mkt, format, nil
}

func RenderGeneric[T any](w io.Writer, format render.Format, res model.Response[T]) (bool, error) {
	switch format {
	case render.FormatJSON:
		return true, renderjson.Render(w, res)
	case render.FormatYAML:
		return true, renderyaml.Render(w, res)
	case render.FormatMarkdown:
		return true, rendermd.Render(w, res)
	case render.FormatToon:
		return true, rendertoon.Render(w, res)
	}
	return false, nil
}
