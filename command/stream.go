package command

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mdnmdn/bits"
	"github.com/mdnmdn/bits/internal/logger"
	"github.com/mdnmdn/bits/model"
	"github.com/mdnmdn/bits/render"
	rendertoon "github.com/mdnmdn/bits/render/toon"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var StreamCmd = &cobra.Command{
	Use:   "stream",
	Short: "Live streaming commands",
}

func init() {
	Root.AddCommand(StreamCmd)
}

var StreamPriceCmd = &cobra.Command{
	Use:   "price <id|symbol>...",
	Short: "Live price feed",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runStreamPrice,
}

var StreamBookCmd = &cobra.Command{
	Use:   "book <id|symbol>...",
	Short: "Live order book feed",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runStreamBook,
}

func init() {
	StreamBookCmd.Flags().Int("depth", 10, "Order book depth (10 or 50)")
	StreamCmd.AddCommand(StreamPriceCmd)
	StreamCmd.AddCommand(StreamBookCmd)
}

func runStreamPrice(cmd *cobra.Command, args []string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	providerID, _, format, err := ResolveOptions(cmd)
	if err != nil {
		return err
	}

	client := bits.NewProvider(cfg, providerID, bits.WithSymbolEngine())

	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	ch, err := client.StartPriceStream(ctx, args)
	if err != nil {
		return err
	}

	for update := range ch {
		if update == nil {
			logger.Default.Debug("stream price: received nil update")
			continue
		}
		logger.Default.Debug("stream price: received update", "price", update.Price, "symbol", update.ID)

		change := "-"
		if update.Change24h != nil {
			change = fmt.Sprintf("%.2f%%", *update.Change24h)
		}

		switch format {
		case render.FormatJSON:
			b, _ := json.Marshal(update)
			_, _ = fmt.Fprintf(os.Stdout, "%s\n", b)

		case render.FormatYAML:
			enc := yaml.NewEncoder(os.Stdout)
			enc.SetIndent(2)
			_ = enc.Encode(update)
			_ = enc.Close()

		case render.FormatMarkdown:
			bidAsk := ""
			if update.BidPrice != nil && update.AskPrice != nil {
				bidAsk = fmt.Sprintf(" | bid:%.4f ask:%.4f", *update.BidPrice, *update.AskPrice)
			}
			vol := ""
			if update.Volume24h != nil {
				vol = fmt.Sprintf(" | vol:%.2f", *update.Volume24h)
			}
			_, _ = fmt.Fprintf(os.Stdout, "- **%s** %.6f %s  _%s_%s%s\n",
				update.ID, update.Price, update.Currency, change, bidAsk, vol)

		case render.FormatToon:
			_ = rendertoon.Render(os.Stdout, model.Response[model.CoinPrice]{
				Kind:     model.KindPrice,
				Provider: client.ID(),
				Data:     *update,
			})

		default:
			bidAsk := ""
			if update.BidPrice != nil && update.AskPrice != nil {
				bidAsk = fmt.Sprintf(" | bid:%.4f ask:%.4f", *update.BidPrice, *update.AskPrice)
			}
			vol := ""
			if update.Volume24h != nil {
				vol = fmt.Sprintf(" | vol:%.2f", *update.Volume24h)
			}
			_, _ = fmt.Fprintf(os.Stdout, "%s  %.6f %s  %s%s%s\n",
				update.ID, update.Price, update.Currency, change, bidAsk, vol)
		}
	}
	return nil
}

func runStreamBook(cmd *cobra.Command, args []string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	providerID, market, format, err := ResolveOptions(cmd)
	if err != nil {
		return err
	}

	client := bits.NewProvider(cfg, providerID, bits.WithSymbolEngine())

	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	depth, _ := cmd.Flags().GetInt("depth")
	ch, err := client.StartOrderBookStream(ctx, args, market, depth)
	if err != nil {
		return err
	}

	for update := range ch {
		if update == nil {
			logger.Default.Debug("stream book: received nil update")
			continue
		}
		logger.Default.Debug("stream book: received update", "symbol", update.Symbol)

		switch format {
		case render.FormatJSON:
			b, _ := json.Marshal(update)
			_, _ = fmt.Fprintf(os.Stdout, "%s\n", b)

		case render.FormatYAML:
			enc := yaml.NewEncoder(os.Stdout)
			enc.SetIndent(2)
			_ = enc.Encode(update)
			_ = enc.Close()

		default:
			_, _ = fmt.Fprintf(os.Stdout, "%s | bids:%d asks:%d\n",
				update.Symbol, len(update.Bids), len(update.Asks))
		}
	}
	return nil
}
