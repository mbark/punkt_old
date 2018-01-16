package cmd

import (
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
}
