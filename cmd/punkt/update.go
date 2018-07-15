package punkt

import (
	"os"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Run update for all managers",
	Long: `Goes through all managers running update for each of them and
also potentially updating their configuration.`,
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
