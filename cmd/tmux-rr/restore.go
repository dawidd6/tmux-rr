package main

import (
	"github.com/dawidd6/tmux-rr/pkg/state"
	"github.com/spf13/cobra"
)

var restoreCommand = &cobra.Command{
	Use:   "restore",
	Short: "Bring back all sessions from state file",
	RunE: func(cmd *cobra.Command, args []string) error {
		return state.Restore()
	},
}
