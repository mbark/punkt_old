package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove a git repository or a file as a symlink",
}

var removeSymlinkCmd = &cobra.Command{
	Use:   "symlink",
	Short: "remove symlink and move file back",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		removeSymlink(cmd, args)
	},
}

var removeGitCmd = &cobra.Command{
	Use:   "repository",
	Short: "remove the given repository from your dotfiles-managed git repos",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		removeGit(cmd, args)
	},
}

func init() {
	removeCmd.AddCommand(removeSymlinkCmd)
	removeCmd.AddCommand(removeGitCmd)
	RootCmd.AddCommand(removeCmd)
}

func removeSymlink(cmd *cobra.Command, args []string) {
	mgr := rootMgr.Symlink()
	err := mgr.Remove(args[0])
	if err != nil {
		logrus.WithError(err).Error("unable to remove symlink")
		os.Exit(1)
	}
}

func removeGit(cmd *cobra.Command, args []string) {
	mgr := rootMgr.Git()
	err := mgr.Remove(args[0])
	if err != nil {
		logrus.WithError(err).Error("unable to remove git repository")
		os.Exit(1)
	}
}
