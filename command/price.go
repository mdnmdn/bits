package command

import (
	"os"

	"github.com/mdnmdn/bits"
	rendertable "github.com/mdnmdn/bits/render/table"
	"github.com/spf13/cobra"
)

var PriceCmd = &cobra.Command{
	Use:   "price <id|symbol>...",
	Short: "Current price(s)",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runPrice,
}

func init() {
	PriceCmd.Flags().String("currency", "usd", "Quote currency")
	Root.AddCommand(PriceCmd)
}

func runPrice(cmd *cobra.Command, args []string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	providerID, market, format, err := ResolveOptions(cmd)
	if err != nil {
		return err
	}

	currency, _ := cmd.Flags().GetString("currency")

	client := bits.NewProvider(cfg, providerID, bits.WithSymbolEngine())

	res, err := client.Price(cmd.Context(), args, currency)
	if err != nil {
		return err
	}
	res.Market = market

	if ok, err := RenderGeneric(os.Stdout, format, res); ok || err != nil {
		return err
	}
	return rendertable.RenderPrices(os.Stdout, res)
}
