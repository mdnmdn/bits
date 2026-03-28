package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/lipgloss"
	"github.com/mdnmdn/bits/internal/capability"
	"github.com/mdnmdn/bits/internal/render"
	"github.com/mdnmdn/bits/internal/resolve"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

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

var (
	streamBookSymbol = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	streamBookBids   = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
	streamBookAsks   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

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
		Provider:   opts.Provider,
		Market:     opts.Market,
		NoFallback: opts.NoFallback,
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
		case render.FormatJSON:
			// JSONL: one compact JSON object per line
			b, _ := json.Marshal(update)
			fmt.Fprintf(os.Stdout, "%s\n", b)

		case render.FormatYAML:
			// one YAML document per update
			enc := yaml.NewEncoder(os.Stdout)
			enc.SetIndent(2)
			_ = enc.Encode(update)
			_ = enc.Close()

		case render.FormatMarkdown:
			// compact markdown line
			fmt.Fprintf(os.Stdout, "- **%s/%s** bids:%d asks:%d\n",
				update.Symbol, update.Market, len(update.Bids), len(update.Asks))

		case render.FormatToon:
			// compact colored line
			fmt.Fprintf(os.Stdout, "%s  bids:%s  asks:%s\n",
				streamBookSymbol.Render(fmt.Sprintf("%s/%s", update.Symbol, update.Market)),
				streamBookBids.Render(fmt.Sprintf("%d", len(update.Bids))),
				streamBookAsks.Render(fmt.Sprintf("%d", len(update.Asks))))

		default:
			fmt.Fprintf(os.Stdout, "[%s/%s] bids:%d asks:%d\n",
				update.Symbol, update.Market, len(update.Bids), len(update.Asks))
		}
	}
	return nil
}
