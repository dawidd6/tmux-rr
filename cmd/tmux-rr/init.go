package main

import (
	"strings"

	"github.com/dawidd6/tmux-rr/pkg/service"
	"github.com/dawidd6/tmux-rr/pkg/shell"
	"github.com/spf13/cobra"
)

var noService bool

var initCommand = &cobra.Command{
	Use:       "init <SHELL>",
	Short:     "Print shell integration code",
	Long:      "Supported shells: " + strings.Join(shell.Supported(), ", "),
	ValidArgs: shell.Supported(),
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if noService {
			return nil
		}
		return service.Run()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return shell.Print(args[0])
	},
}

func init() {
	initCommand.Flags().BoolVar(&noService, "no-service", false, "Don't install service")
}
