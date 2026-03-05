package cmd

import (
	"fmt"
	"strings"

	"coingecko-cli/internal/config"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Configure API key and tier",
	RunE:  runAuth,
}

func init() {
	authCmd.Flags().String("key", "", "CoinGecko API key")
	authCmd.Flags().String("tier", "", "API tier (demo, analyst, lite, pro, enterprise)")
	rootCmd.AddCommand(authCmd)
}

func runAuth(cmd *cobra.Command, args []string) error {
	key, _ := cmd.Flags().GetString("key")
	tier, _ := cmd.Flags().GetString("tier")

	if key == "" {
		if err := huh.NewInput().
			Title("API Key").
			Description("Enter your CoinGecko API key").
			Value(&key).
			Run(); err != nil {
			return err
		}
	}

	if tier == "" {
		if err := huh.NewSelect[string]().
			Title("API Tier").
			Options(
				huh.NewOption("Demo (free)", config.TierDemo),
				huh.NewOption("Analyst", config.TierAnalyst),
				huh.NewOption("Lite", config.TierLite),
				huh.NewOption("Pro", config.TierPro),
				huh.NewOption("Enterprise", config.TierEnterprise),
			).
			Value(&tier).
			Run(); err != nil {
			return err
		}
	}

	tier = strings.ToLower(tier)
	if !config.IsValidTier(tier) {
		return fmt.Errorf("invalid tier %q — must be one of: %s", tier, strings.Join(config.ValidTiers, ", "))
	}

	cfg := &config.Config{APIKey: key, Tier: tier}
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("Saved! Tier: %s, Key: %s\n", cfg.Tier, cfg.MaskedKey())
	return nil
}
