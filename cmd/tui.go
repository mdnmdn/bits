package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mdnmdn/bits/internal/tui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI",
	RunE:  runTUI,
}

var tuiSection string
var tuiRefresh string

func init() {
	RootCmd.AddCommand(tuiCmd)

	tuiCmd.Flags().StringVarP(&tuiSection, "section", "s", "", "Section to open (prices, ticker, exchange, book, candles, markets, trending)")
	tuiCmd.Flags().StringVarP(&tuiRefresh, "refresh", "r", "", "Auto-refresh interval (5s, 10s, 30s, 1m)")

	tuiCmd.AddCommand(&cobra.Command{
		Use:   "prices",
		Short: "Prices section",
		RunE:  runTUISection("prices"),
	})
	tuiCmd.AddCommand(&cobra.Command{
		Use:   "ticker",
		Short: "Ticker section",
		RunE:  runTUISection("ticker"),
	})
	tuiCmd.AddCommand(&cobra.Command{
		Use:   "exchange",
		Short: "Exchange section",
		RunE:  runTUISection("exchange"),
	})
	tuiCmd.AddCommand(&cobra.Command{
		Use:   "book",
		Short: "Order book section",
		RunE:  runTUISection("book"),
	})
	tuiCmd.AddCommand(&cobra.Command{
		Use:   "candles",
		Short: "Candles section",
		RunE:  runTUISection("candles"),
	})
	tuiCmd.AddCommand(&cobra.Command{
		Use:   "markets",
		Short: "Markets section",
		RunE:  runTUISection("markets"),
	})
	tuiCmd.AddCommand(&cobra.Command{
		Use:   "trending",
		Short: "Trending section",
		RunE:  runTUISection("trending"),
	})
}

func runTUI(cmd *cobra.Command, args []string) error {
	return runTUISection(tuiSection)(cmd, args)
}

func runTUISection(section string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		opts := tui.Options{
			Section:         section,
			Provider:        getProviderFlag(cmd),
			Market:          getMarketFlag(cmd),
			Symbol:          getSymbolFlag(cmd),
			RefreshInterval: tuiRefresh,
		}

		p := tea.NewProgram(tui.NewApp(opts), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return err
		}
		return nil
	}
}

func getProviderFlag(cmd *cobra.Command) string {
	if cmd.Flags().Changed("provider") {
		p, _ := cmd.Flags().GetString("provider")
		return p
	}
	return ""
}

func getMarketFlag(cmd *cobra.Command) string {
	if cmd.Flags().Changed("market") {
		m, _ := cmd.Flags().GetString("market")
		return m
	}
	return ""
}

func getSymbolFlag(cmd *cobra.Command) string {
	if cmd.Flags().Changed("symbol") {
		s, _ := cmd.Flags().GetString("symbol")
		return s
	}
	return ""
}
