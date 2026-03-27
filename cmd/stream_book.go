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

var streamBookCmd = &cobra.Command{
	Use:   "book <symbol>",
	Short: "Live order book feed",
	Args:  cobra.ExactArgs(1),
	RunE:  runStreamBook,
}

func init() {
	streamBookCmd.Flags().Int("depth", 20, "Order book depth")
	streamCmd.AddCommand(streamBookCmd)
}

func runStreamBook(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	opts := resolveOpts(cmd)
	format := resolveFormat(cmd)
	depth, _ := cmd.Flags().GetInt("depth")
	resolver := newResolver(cfg)

	p, market, _, rerr := resolver.Resolve(cmd.Context(), capability.FeatureStreamOrderBook, resolve.ResolutionOpts{
		Provider: opts.Provider,
		Market:   opts.Market,
		Lock:     opts.Lock,
	})
	if rerr != nil {
		return rerr
	}

	obsp, rerr := resolve.Require[provider.OrderBookStreamProvider](p, "stream-book")
	if rerr != nil {
		return rerr
	}

	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	ch, err := obsp.WatchOrderBook(ctx, args[0], market, depth)
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
			fmt.Fprintf(os.Stdout, "[%s/%s] bids:%d asks:%d\n",
				update.Symbol, update.Market, len(update.Bids), len(update.Asks))
		}
	}
	return nil
}
