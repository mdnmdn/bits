package cmd

import (
	"fmt"

	"github.com/mdnmdn/bits/internal/capability"
	"github.com/mdnmdn/bits/internal/display"
	"github.com/mdnmdn/bits/internal/provider"
	"github.com/spf13/cobra"
)

var capabilitiesCmd = &cobra.Command{
	Use:     "capabilities",
	Aliases: []string{"caps"},
	Short:   "Show provider capability matrix",
	Long:    "Display which features are supported by each provider and market type.",
	Example: `  bits capabilities
  bits caps -p binance
  bits capabilities -o json`,
	RunE: runCapabilities,
}

func init() {
	RootCmd.AddCommand(capabilitiesCmd)
}

type capabilityRow struct {
	Feature   string          `json:"feature"`
	Market    string          `json:"market"`
	Providers map[string]bool `json:"providers"`
}

func runCapabilities(cmd *cobra.Command, args []string) error {
	providerFilter, _ := cmd.Root().PersistentFlags().GetString("provider")
	jsonOut := outputJSON(cmd)

	allCaps := provider.AllCapabilities()

	providerNames := provider.AvailableProviders
	if providerFilter != "" {
		if _, ok := allCaps[providerFilter]; !ok {
			return fmt.Errorf("unknown provider %q", providerFilter)
		}
		providerNames = []string{providerFilter}
	}

	if jsonOut {
		return printJSONRaw(buildCapabilityJSON(allCaps, providerNames))
	}

	display.PrintBanner()
	headers, rows := buildCapabilityTable(allCaps, providerNames)
	display.PrintTable(headers, rows)
	return nil
}

func buildCapabilityTable(allCaps map[string]capability.CapabilityMatrix, providerNames []string) ([]string, [][]string) {
	check := "✓"
	dash := "-"
	if display.StdoutColorEnabled() {
		check = "\033[32m✓\033[0m"
	}

	headers := append([]string{"Feature", "Market"}, providerNames...)
	var rows [][]string

	for _, feat := range capability.AllFeatures() {
		for _, mkt := range capability.AllMarkets() {
			key := capability.CapabilityKey{Market: mkt, Feature: feat}
			// skip row if no provider supports this combination
			anySupported := false
			for _, name := range providerNames {
				if allCaps[name][key] {
					anySupported = true
					break
				}
			}
			if !anySupported {
				continue
			}

			row := []string{string(feat), string(mkt)}
			for _, name := range providerNames {
				if allCaps[name][key] {
					row = append(row, check)
				} else {
					row = append(row, dash)
				}
			}
			rows = append(rows, row)
		}
	}
	return headers, rows
}

func buildCapabilityJSON(allCaps map[string]capability.CapabilityMatrix, providerNames []string) []capabilityRow {
	var result []capabilityRow
	for _, feat := range capability.AllFeatures() {
		for _, mkt := range capability.AllMarkets() {
			key := capability.CapabilityKey{Market: mkt, Feature: feat}
			providers := make(map[string]bool, len(providerNames))
			anySupported := false
			for _, name := range providerNames {
				v := allCaps[name][key]
				providers[name] = v
				if v {
					anySupported = true
				}
			}
			if !anySupported {
				continue
			}
			result = append(result, capabilityRow{
				Feature:   string(feat),
				Market:    string(mkt),
				Providers: providers,
			})
		}
	}
	return result
}
