package command

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/mdnmdn/bits/capability"
	"github.com/mdnmdn/bits/provider/registry"
	"github.com/spf13/cobra"
)

var ProvidersCmd = &cobra.Command{
	Use:   "providers",
	Short: "List registered providers",
	RunE:  runProviders,
}

func init() {
	Root.AddCommand(ProvidersCmd)
}

func runProviders(cmd *cobra.Command, args []string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	ids := registry.AllProviderIDs()
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "PROVIDER\tACTIVE")
	for _, id := range ids {
		active := ""
		if id == cfg.ActiveProvider() {
			active = "*"
		}
		_, _ = fmt.Fprintf(tw, "%s\t%s\n", id, active)
	}
	return tw.Flush()
}

var CapabilitiesCmd = &cobra.Command{
	Use:     "capabilities",
	Aliases: []string{"caps"},
	Short:   "Show provider capability matrix",
	RunE:    runCapabilities,
}

func init() {
	Root.AddCommand(CapabilitiesCmd)
}

func runCapabilities(cmd *cobra.Command, args []string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	providerFilter, _ := cmd.Root().PersistentFlags().GetString("provider")
	ids := registry.AllProviderIDs()

	allCaps := make(map[string]capability.CapabilityMatrix, len(ids))
	for _, id := range ids {
		p, perr := registry.NewProvider(id, cfg)
		if perr != nil {
			continue
		}
		allCaps[id] = p.Capabilities()
	}

	names := ids
	if providerFilter != "" {
		if _, ok := allCaps[providerFilter]; !ok {
			return fmt.Errorf("unknown provider %q", providerFilter)
		}
		names = []string{providerFilter}
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	header := "FEATURE\tMARKET"
	for _, n := range names {
		header += "\t" + n
	}
	_, _ = fmt.Fprintln(tw, header)

	for _, feat := range capability.AllFeatures() {
		for _, mkt := range capability.AllMarkets() {
			key := capability.CapabilityKey{Market: mkt, Feature: feat}
			anySupported := false
			for _, n := range names {
				if allCaps[n][key] {
					anySupported = true
					break
				}
			}
			if !anySupported {
				continue
			}
			row := fmt.Sprintf("%s\t%s", feat, mkt)
			for _, n := range names {
				if allCaps[n][key] {
					row += "\t✓"
				} else {
					row += "\t-"
				}
			}
			_, _ = fmt.Fprintln(tw, row)
		}
	}
	return tw.Flush()
}
