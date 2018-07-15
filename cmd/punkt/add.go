package punkt

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "add a symlink or repository",
}

var addSymlinkCmd = &cobra.Command{
	Use:   "symlink target [new-location]",
	Short: "Create a symlink from target to a new location and store it",
	Long: `Create a symlink from target to a new location, which is optional and will be default
be inferred, and save the symlink to your configured symlinks. To undo the operation
see remove.

If you don't specify a new location the new location will default to having the same
relative path to your dotfiles directory as it currently has to your home directory
(i.e placing ~/.config/git/ignore in ~/dotfiles/.config/git/ignore).`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		addSymlink(cmd, args)
	},
}

var addGitCmd = &cobra.Command{
	Use:   "repository path",
	Short: "Add the git repository to the dotfile git configuration",
	Long:  `Add the target git repository to the configuration file for git repositories.`,
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
