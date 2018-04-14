package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update all packages",
	Long:  `update all package versions`,
	Run: func(cmd *cobra.Command, args []string) {
		update()
	},
}

func init() {
	RootCmd.AddCommand(updateCmd)
}

// Update ...
func update() {
	for i := range managers {
		_, err := managers[i].Update()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"manager": managers[i],
			}).WithError(err).Error("Command ensure failed for manager")
		}
	}
}
