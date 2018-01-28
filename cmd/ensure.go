package cmd

import (
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
	for i := range managers {
		err := managers[i].Ensure()
		if err != nil {
			logrus.WithField("manager", managers[i]).WithError(err).Error("Command ensure failed for manager")
		}
	}
}
