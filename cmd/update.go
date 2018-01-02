package cmd

import (
	"os/user"

	"github.com/mbark/punkt/exec"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update all packages",
	Long:  `update all package versions`,
	Run: func(cmd *cobra.Command, args []string) {
		Update()
	},
}

func init() {
	RootCmd.AddCommand(updateCmd)
}

// Update ...
func Update() {
	usr, err := user.Current()
	if err != nil {
		logrus.WithError(err).Fatal("Unable to get current user")
	}

	exec.Run("ansible-playbook", "main.yml", "-i", "inventory", "--become-user="+usr.Username, "-e", "punkt_upgrade=true")
}
