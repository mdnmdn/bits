package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/lipgloss"
	"github.com/mdnmdn/bits/internal/render"
	renderjson "github.com/mdnmdn/bits/internal/render/json"
	rendertoon "github.com/mdnmdn/bits/internal/render/toon"
	"github.com/mdnmdn/bits/pkg/capability"
	"github.com/mdnmdn/bits/pkg/model"
	"github.com/mdnmdn/bits/pkg/resolve"
	"github.com/mdnmdn/bits/pkg/resolve/symbol"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/mdnmdn/bits/pkg/provider"
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

	sym, _ := symbol.New(p).Resolve(ctx, args[0], market) // falls back to raw input on error
	if sym == "" {
		sym = args[0]
	}
	ch, err := obsp.StartOrderBookStream(ctx, []string{sym}, market, depth)
	if err != nil {
		return err
	}

	for update := range ch {
		if update == nil {
			continue
		}

		res := model.Response[model.OrderBook]{
			Kind:     model.KindOrderBook,
			Provider: p.ID(),
			Market:   market,
			Data:     *update,
		}

		switch format {
		case render.FormatJSON:
			_ = renderjson.Render(os.Stdout, res)

		case render.FormatYAML:
			enc := yaml.NewEncoder(os.Stdout)
			enc.SetIndent(2)
			_ = enc.Encode(res)
			_ = enc.Close()

		case render.FormatMarkdown:
			_, _ = fmt.Fprintf(os.Stdout, "- **%s/%s** bids:%d asks:%d\n",
				update.Symbol, update.Market, len(update.Bids), len(update.Asks))

		case render.FormatToon:
			_ = rendertoon.RenderOrderBook(os.Stdout, res)

		default:
			// Default: show top of book with actual values
			_, _ = fmt.Fprintf(os.Stdout, "[%s]\n", streamBookSymbol.Render(fmt.Sprintf("%s/%s", update.Symbol, update.Market)))
			if len(update.Bids) > 0 {
				bid := update.Bids[0]
				_, _ = fmt.Fprintf(os.Stdout, "  %sbids:  %s @ %s\n",
					streamBookBids.Render(""),
					streamBookBids.Render(fmt.Sprintf("%.4f", bid.Quantity)),
					streamBookBids.Render(fmt.Sprintf("%.2f", bid.Price)))
			}
			if len(update.Asks) > 0 {
				ask := update.Asks[0]
				_, _ = fmt.Fprintf(os.Stdout, "  %sasks:  %s @ %s\n",
					streamBookAsks.Render(""),
					streamBookAsks.Render(fmt.Sprintf("%.4f", ask.Quantity)),
					streamBookAsks.Render(fmt.Sprintf("%.2f", ask.Price)))
			}
		}
	}
	return nil
}
