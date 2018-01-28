package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

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
	from := args[0]
	to := ""
	if len(args) > 1 {
		to = args[1]
	}

	mgr := symlink.NewManager(*config)
	err := mgr.Add(args[0], to)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"from": from,
			"to":   to,
		}).WithError(err).Error("Failed to create symlink")
	}
}
