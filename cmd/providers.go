package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/mdnmdn/bits/provider/registry"
	"github.com/spf13/cobra"
)

var providersCmd = &cobra.Command{
	Use:   "providers",
	Short: "List registered providers",
	RunE:  runProviders,
}

func init() {
	RootCmd.AddCommand(providersCmd)
}

func runProviders(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
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
