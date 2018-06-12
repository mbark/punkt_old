package punkt

import (
	"os"

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
	err := rootMgr.Ensure(rootMgr.All())
	if err != nil {
		os.Exit(1)
	}
}
