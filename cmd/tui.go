package cmd

import (
	"fmt"

	"github.com/mdnmdn/bits/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Interactive TUI for browsing crypto data",
	Example: `  cg tui markets
  cg tui trending
  cg tui markets --category layer-1`,
}

var tuiMarketsCmd = &cobra.Command{
	Use:   "markets",
	Short: "Browse top coins by market cap",
	Example: `  cg tui markets
  cg tui markets --vs eur --category layer-1`,
	RunE: runTUIMarkets,
}

var tuiTrendingCmd = &cobra.Command{
	Use:   "trending",
	Short: "Browse trending coins",
	Example: `  cg tui trending
  cg tui trending --vs eur`,
	RunE: runTUITrending,
}

func init() {
	tuiMarketsCmd.Flags().Int("total", 50, "Total number of coins to fetch")
	tuiMarketsCmd.Flags().String("vs", "usd", "Target currency")
	tuiMarketsCmd.Flags().String("category", "", "Filter by category")
	tuiTrendingCmd.Flags().String("vs", "usd", "Target currency")

	tuiCmd.AddCommand(tuiMarketsCmd)
	tuiCmd.AddCommand(tuiTrendingCmd)
	RootCmd.AddCommand(tuiCmd)
}

func runTUIMarkets(cmd *cobra.Command, args []string) error {
	total, _ := cmd.Flags().GetInt("total")
	vs, _ := cmd.Flags().GetString("vs")
	category, _ := cmd.Flags().GetString("category")

	if total <= 0 {
		return fmt.Errorf("--total must be a positive integer")
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	client := newAPIClient(cfg)
	model := tui.NewMarketsModel(client, vs, category, total)

	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func runTUITrending(cmd *cobra.Command, args []string) error {
	vs, _ := cmd.Flags().GetString("vs")

	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	client := newAPIClient(cfg)
	showMax := ""
	if cfg.IsPaid() {
		showMax = "coins"
	}
	model := tui.NewTrendingModel(client, vs, showMax)

	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
