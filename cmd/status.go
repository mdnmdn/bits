package cmd

import (
	"fmt"

	"coingecko-cli/internal/config"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current configuration",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if outputJSON(cmd) {
		return printJSONRaw(map[string]string{
			"tier":     cfg.Tier,
			"api_key":  cfg.MaskedKey(),
			"base_url": cfg.BaseURL(),
		})
	}

	fmt.Printf("Tier:     %s\n", cfg.Tier)
	fmt.Printf("API Key:  %s\n", cfg.MaskedKey())
	fmt.Printf("Base URL: %s\n", cfg.BaseURL())
	return nil
}
