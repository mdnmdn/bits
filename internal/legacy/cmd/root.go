package legacycmd

import (
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var LegacyCmd = &cobra.Command{
	Use:    "legacy",
	Short:  "Legacy commands (kept during transition)",
	Long:   "A command-line tool for accessing multi-provider cryptocurrency market data.",
	Hidden: true,
}

func init() {
}

func outputJSON(cmd *cobra.Command) bool {
	o, _ := cmd.Flags().GetString("output")
	return o == "json"
}
