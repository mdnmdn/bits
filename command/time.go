package command

import (
	"os"

	"github.com/mdnmdn/bits"
	rendertable "github.com/mdnmdn/bits/render/table"
	"github.com/spf13/cobra"
)

var TimeCmd = &cobra.Command{
	Use:   "time",
	Short: "Show exchange server time",
	RunE:  runTime,
}

func init() {
	Root.AddCommand(TimeCmd)
}

func runTime(cmd *cobra.Command, args []string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	providerID, market, format, err := ResolveOptions(cmd)
	if err != nil {
		return err
	}

	client := bits.NewProvider(cfg, providerID, bits.WithSymbolEngine())

	res, err := client.ServerTime(cmd.Context())
	if err != nil {
		return err
	}
	res.Market = market

	if ok, err := RenderGeneric(os.Stdout, format, res); ok || err != nil {
		return err
	}
	return rendertable.RenderServerTime(os.Stdout, res)
}
