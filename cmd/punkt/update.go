package punkt

import (
	"os"

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
	err := rootMgr.Update(rootMgr.All())
	if err != nil {
		os.Exit(1)
	}
}
