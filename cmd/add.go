package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/mbark/punkt/file"
	"github.com/mbark/punkt/symlink"
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

	created := symlink.Add(args[0], to, dotfiles)
	if created == nil {
		return
	}

	symlinks := []symlink.Symlink{}
	file.Read(&symlinks, dotfiles, "symlinks")

	for _, s := range symlinks {
		if s == *created {
			logrus.WithField("symlink", created).Info("Symlink already stored in file, not adding")
			return
		}
	}

	symlinks = append(symlinks, *created)

	file.Save(symlinks, dotfiles, "symlinks")
}
