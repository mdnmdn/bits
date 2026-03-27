package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mdnmdn/bits/internal/capability"
	"github.com/mdnmdn/bits/internal/resolve"
	"github.com/spf13/cobra"

	"github.com/mdnmdn/bits/internal/provider"
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
		Provider: opts.Provider,
		Market:   opts.Market,
		Lock:     opts.Lock,
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

	ch, err := sp.WatchPrices(ctx, args)
	if err != nil {
		return err
	}

	for update := range ch {
		if update == nil {
			continue
		}
		switch format {
		case "json":
			b, _ := json.Marshal(update)
			fmt.Fprintf(os.Stdout, "%s\n", b)
		default:
			change := "-"
			if update.Change24h != nil {
				change = fmt.Sprintf("%.2f%%", *update.Change24h)
			}
			fmt.Fprintf(os.Stdout, "%s  %.6f %s  %s\n", update.ID, update.Price, update.Currency, change)
		}
	}
	return nil
}
