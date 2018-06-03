package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "add a git repository or a file as a symlink",
}

var addSymlinkCmd = &cobra.Command{
	Use:   "symlink target [new-location]",
	Short: "store target in dotfile's directory and link to it",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		addSymlink(cmd, args)
	},
}

var addGitCmd = &cobra.Command{
	Use:   "repository path",
	Short: "add the given repository to your dotfiles-managed git repos",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		addGit(cmd, args)
	},
}

func init() {
	addCmd.AddCommand(addSymlinkCmd)
	addCmd.AddCommand(addGitCmd)
	RootCmd.AddCommand(addCmd)
}

func addSymlink(cmd *cobra.Command, args []string) {
	newLocation := ""
	if len(args) == 2 {
		newLocation = args[1]
	}

	mgr := rootMgr.Symlink()
	_, err := mgr.Add(args[0], newLocation)
	if err != nil {
		logrus.WithError(err).Error("failed to add symlink")
		os.Exit(1)
	}
}

func addGit(cmd *cobra.Command, args []string) {
	mgr := rootMgr.Git()
	err := mgr.Add(args[0])
	if err != nil {
		logrus.WithError(err).Error("failed to add git repo")
		os.Exit(1)
	}
}
