package legacycmd

import (
	"fmt"

	"github.com/mdnmdn/bits/internal/legacy/display"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current configuration",
	RunE:  runStatus,
}

func init() {
	LegacyCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if outputJSON(cmd) {
		return printJSONRaw(map[string]string{
			"tier":     cfg.CoinGecko.Tier,
			"api_key":  cfg.MaskedKey(),
			"base_url": cfg.BaseURL(),
		})
	}

	display.PrintBanner()
	fmt.Printf("Tier:     %s\n", cfg.CoinGecko.Tier)
	fmt.Printf("API Key:  %s\n", cfg.MaskedKey())
	fmt.Printf("Base URL: %s\n", cfg.BaseURL())
	return nil
}
