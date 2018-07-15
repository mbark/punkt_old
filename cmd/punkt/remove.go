package punkt

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove a {symlink,repository} from your dotfiles",
}

var removeSymlinkCmd = &cobra.Command{
	Use:   "symlink",
	Short: "Remove the symlink from your dotfiles and put the file back",
	Long: `Remove the symlink form your dotfiles' symlink configuration file,
removes the symlik and moves the file back to its original position.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		removeSymlink(cmd, args)
	},
}

var removeGitCmd = &cobra.Command{
	Use:   "repository",
	Short: "Remove the git repository from your dotfiles",
	Long:  `Remove the git repository from your dotfiles' git configuration file`,
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
