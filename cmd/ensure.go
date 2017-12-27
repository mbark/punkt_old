package cmd

import (
	"os"
	"os/user"

	"github.com/mbark/punkt/exec"
	"github.com/mbark/punkt/path"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ensureCmd = &cobra.Command{
	Use:   "ensure",
	Short: "Ensure your dotfiles are up to date",
	Long: `Ensure that your dotfiles are up to date with your
current environment.`,
	Run: func(cmd *cobra.Command, args []string) {
		ensure()
	},
}

func init() {
	RootCmd.AddCommand(ensureCmd)
}

func ensure() {
	path.GoToPunktHome()

	usr, err := user.Current()
	if err != nil {
		logrus.WithError(err).Fatal("Unable to get current user")
	}

	os.Chdir("ansible")
	exec.Run("ansible-playbook", "main.yml", "-i", "inventory", "--become-user="+usr.Username)
}
