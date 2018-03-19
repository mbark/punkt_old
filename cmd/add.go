package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/mbark/punkt/mgr/symlink"
)

// ensureCmd represents the ensure command
var addCmd = &cobra.Command{
	Use:   "add from [to]",
	Short: "add file as a symlink to your dotfiles",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		add(cmd, args)
	},
}

func init() {
	RootCmd.AddCommand(addCmd)
}

func add(cmd *cobra.Command, args []string) {
	mgr := symlink.NewManager(*config)
	link, err := mgr.Add(args[0])
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"from": link.From,
			"to":   link.To,
		}).WithError(err).Error("Failed to create symlink")
		os.Exit(1)
	}
}
