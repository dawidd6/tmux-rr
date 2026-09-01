package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type Session struct{}

func main() {
	command := &cobra.Command{
		Use:           "tmux-rr",
		Short:         "Save and restore all tmux sessions automatically.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	command.AddCommand(
		clearCommand,
		daemonCommand,
		initCommand,
		restoreCommand,
		saveCommand,
	)

	if err := command.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
