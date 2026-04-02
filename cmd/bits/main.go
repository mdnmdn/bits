package main

import (
	"fmt"
	"os"

	"github.com/mdnmdn/bits/command"
)

func main() {
	command.Root.SilenceUsage = true
	command.Root.SilenceErrors = true

	if err := command.Root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
