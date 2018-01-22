package cmd

import (
	"github.com/spf13/cobra"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/mgr/symlink"
)

// ensureCmd represents the ensure command
var addCmd = &cobra.Command{
	Use:   "add from [to]",
	Short: "add file as a symlink to your dotfiles",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		add(cmd, args)
	},
}

func init() {
	RootCmd.AddCommand(addCmd)
}

func add(cmd *cobra.Command, args []string) {
	to := ""
	if len(args) > 1 {
		to = args[1]
	}

	mgr := symlink.NewManager(conf.Config{
		Dotfiles:  dotfiles,
		PunktHome: punktHome,
	})
	mgr.Add(args[0], to)
}
