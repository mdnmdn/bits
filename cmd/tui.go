package cmd

import (
	"coingecko-cli/internal/api"
	"coingecko-cli/internal/config"
	"coingecko-cli/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Interactive TUI for browsing markets",
	RunE:  runTUI,
}

func init() {
	tuiCmd.Flags().String("vs", "usd", "Target currency")
	tuiCmd.Flags().String("category", "", "Filter by category")
	rootCmd.AddCommand(tuiCmd)
}

func runTUI(cmd *cobra.Command, args []string) error {
	vs, _ := cmd.Flags().GetString("vs")
	category, _ := cmd.Flags().GetString("category")

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	client := api.NewClient(cfg)
	model := tui.NewMarketsModel(client, vs, category)

	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
