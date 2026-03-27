package cmd

import "github.com/spf13/cobra"

var streamCmd = &cobra.Command{
	Use:   "stream",
	Short: "Live streaming commands",
}

func init() {
	RootCmd.AddCommand(streamCmd)
}
