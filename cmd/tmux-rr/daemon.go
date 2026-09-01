package main

import (
	"errors"
	"os/signal"
	"syscall"

	"github.com/dawidd6/tmux-rr/pkg/state"
	"github.com/dawidd6/tmux-rr/pkg/tmux"
	"github.com/spf13/cobra"
)

var daemonCommand = &cobra.Command{
	Use:    "daemon",
	Short:  "Start the daemon",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		if err = tmux.StartServer(); err != nil {
			return err
		}

		defer func() {
			err = errors.Join(err, tmux.KillServer())
		}()

		if err = state.Restore(); err != nil {
			return err
		}

		<-ctx.Done()
		stop()

		if err = state.Save(); err != nil {
			return err
		}

		return nil
	},
}
