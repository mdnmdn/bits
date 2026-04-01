package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mdnmdn/bits/capability"
	"github.com/mdnmdn/bits/internal/render"
	rendertoon "github.com/mdnmdn/bits/internal/render/toon"
	"github.com/mdnmdn/bits/model"
	"github.com/mdnmdn/bits/resolve"
	"github.com/mdnmdn/bits/resolve/symbol"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/mdnmdn/bits/provider"
)

var streamPriceCmd = &cobra.Command{
	Use:   "price <id|symbol>...",
	Short: "Live price feed",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runStreamPrice,
}

func init() {
	streamCmd.AddCommand(streamPriceCmd)
}

func runStreamPrice(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	opts := resolveOpts(cmd)
	format := resolveFormat(cmd)
	resolver := newResolver(cfg)

	p, _, _, rerr := resolver.Resolve(cmd.Context(), capability.FeatureStreamPrice, resolve.ResolutionOpts{
		Provider:   opts.Provider,
		Market:     opts.Market,
		NoFallback: opts.NoFallback,
	})
	if rerr != nil {
		return rerr
	}

	sp, rerr := resolve.Require[provider.PriceStreamProvider](p, "stream-price")
	if rerr != nil {
		return rerr
	}

	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	ids := make([]string, len(args))
	for i, arg := range args {
		sym, _ := symbol.New(p).Resolve(ctx, arg, opts.Market)
		if sym == "" {
			sym = arg
		}
		ids[i] = sym
	}

	ch, err := sp.StartPriceStream(ctx, ids)
	if err != nil {
		return err
	}

	for update := range ch {
		if update == nil {
			continue
		}
		change := "-"
		if update.Change24h != nil {
			change = fmt.Sprintf("%.2f%%", *update.Change24h)
		}

		switch format {
		case render.FormatJSON:
			// JSONL: one compact JSON object per line
			b, _ := json.Marshal(update)
			_, _ = fmt.Fprintf(os.Stdout, "%s\n", b)

		case render.FormatYAML:
			// one YAML document per update, separated by ---
			enc := yaml.NewEncoder(os.Stdout)
			enc.SetIndent(2)
			_ = enc.Encode(update)
			_ = enc.Close()

		case render.FormatMarkdown:
			// one markdown bullet per update with bid/ask and volume
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
			res := model.Response[model.CoinPrice]{
				Kind:     model.KindPrice,
				Provider: p.ID(),
				Data:     *update,
			}
			_ = rendertoon.RenderPrice(os.Stdout, res)

		default:
			// table: compact plain line with bid/ask and volume
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
