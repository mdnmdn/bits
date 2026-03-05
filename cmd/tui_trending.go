package cmd

import (
	"coingecko-cli/internal/api"
	"coingecko-cli/internal/config"
	"coingecko-cli/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var tuiTrendingCmd = &cobra.Command{
	Use:   "tui-trending",
	Short: "Interactive TUI for browsing trending coins",
	RunE:  runTUITrending,
}

func init() {
	tuiTrendingCmd.Flags().String("vs", "usd", "Target currency")
	rootCmd.AddCommand(tuiTrendingCmd)
}

func runTUITrending(cmd *cobra.Command, args []string) error {
	vs, _ := cmd.Flags().GetString("vs")

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	client := api.NewClient(cfg)
	model := tui.NewTrendingModel(client, vs)

	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
