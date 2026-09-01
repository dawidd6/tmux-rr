package main

import (
	"github.com/dawidd6/tmux-rr/pkg/state"
	"github.com/spf13/cobra"
)

var saveCommand = &cobra.Command{
	Use:   "save",
	Short: "Remember all sessions in state file",
	RunE: func(cmd *cobra.Command, args []string) error {
		return state.Save()
	},
}
