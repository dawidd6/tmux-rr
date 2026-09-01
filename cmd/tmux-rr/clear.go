package main

import (
	"github.com/dawidd6/tmux-rr/pkg/state"
	"github.com/spf13/cobra"
)

var clearCommand = &cobra.Command{
	Use:   "clear",
	Short: "Forget sessions from state file",
	RunE: func(cmd *cobra.Command, args []string) error {
		return state.Clear()
	},
}
