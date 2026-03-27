package cmd

import (
	"fmt"

	"github.com/mdnmdn/bits/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration (with API keys redacted)",
	RunE:  runConfigShow,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a new configuration file",
	RunE:  runConfigInit,
}

var initLocal bool

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configInitCmd)
	configInitCmd.Flags().BoolVarP(&initLocal, "local", "l", false, "Create config in current directory")
	rootCmd.AddCommand(configCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	redacted := cfg.Redacted()

	if outputJSON(cmd) {
		return printJSONRaw(redacted)
	}

	fmt.Printf("Config: %s\n\n", loadedConfigPath)
	fmt.Printf("Provider: %s\n\n", redacted.Provider)

	fmt.Println("CoinGecko:")
	fmt.Printf("  Tier:    %s\n", redacted.CoinGecko.Tier)
	fmt.Printf("  API Key: %s\n", redacted.CoinGecko.APIKey)
	fmt.Printf("  Base URL: %s\n\n", redacted.CoinGecko.BaseURL)

	fmt.Println("Binance:")
	fmt.Printf("  API Key:    %s\n", redacted.Binance.APIKey)
	fmt.Printf("  API Secret: %s\n", redacted.Binance.APISecret)
	fmt.Printf("  Base URL:   %s\n", redacted.Binance.BaseURL)
	fmt.Printf("  Spot:    Enabled=%v\n", redacted.Binance.Spot.Enabled)
	fmt.Printf("  Margin:  Enabled=%v\n", redacted.Binance.Margin.Enabled)
	fmt.Printf("  Futures: Enabled=%v, UseTestnet=%v\n\n", redacted.Binance.Futures.Enabled, redacted.Binance.Futures.UseTestnet)

	fmt.Println("Bitget:")
	fmt.Printf("  API Key:     %s\n", redacted.Bitget.APIKey)
	fmt.Printf("  API Secret:  %s\n", redacted.Bitget.APISecret)
	fmt.Printf("  Passphrase:  %s\n", redacted.Bitget.Passphrase)
	fmt.Printf("  Base URL:    %s\n", redacted.Bitget.BaseURL)
	fmt.Printf("  Spot:    Enabled=%v\n", redacted.Bitget.Spot.Enabled)
	fmt.Printf("  Futures: Enabled=%v\n", redacted.Bitget.Futures.Enabled)

	return nil
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	path, err := config.Init(initLocal)
	if err != nil {
		return err
	}
	fmt.Printf("Config file created: %s\n", path)
	return nil
}
