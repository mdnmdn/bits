package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/lipgloss"
	"github.com/mdnmdn/bits/internal/render"
	"github.com/mdnmdn/bits/pkg/capability"
	"github.com/mdnmdn/bits/pkg/resolve"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/mdnmdn/bits/pkg/provider"
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

var (
	streamToonID     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	streamToonPrice  = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	streamToonUp     = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
	streamToonDown   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	streamToonNeutrl = lipgloss.NewStyle().Faint(true)
)

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

	ch, err := sp.WatchPrices(ctx, args)
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
			fmt.Fprintf(os.Stdout, "%s\n", b)

		case render.FormatYAML:
			// one YAML document per update, separated by ---
			enc := yaml.NewEncoder(os.Stdout)
			enc.SetIndent(2)
			_ = enc.Encode(update)
			_ = enc.Close()

		case render.FormatMarkdown:
			// one markdown bullet per update
			fmt.Fprintf(os.Stdout, "- **%s** %.6f %s  _%s_\n",
				update.ID, update.Price, update.Currency, change)

		case render.FormatToon:
			// compact colored line: ID  PRICE CURRENCY  CHANGE
			chgStyle := streamToonNeutrl
			if update.Change24h != nil {
				if *update.Change24h > 0 {
					change = "▲" + change
					chgStyle = streamToonUp
				} else if *update.Change24h < 0 {
					change = "▼" + change
					chgStyle = streamToonDown
				}
			}
			fmt.Fprintf(os.Stdout, "%s  %s %s  %s\n",
				streamToonID.Render(update.ID),
				streamToonPrice.Render(fmt.Sprintf("%.6f", update.Price)),
				update.Currency,
				chgStyle.Render(change))

		default:
			// table: compact plain line
			fmt.Fprintf(os.Stdout, "%s  %.6f %s  %s\n",
				update.ID, update.Price, update.Currency, change)
		}
	}
	return nil
}
