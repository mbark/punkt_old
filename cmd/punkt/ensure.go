package punkt

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var ensureLongMsg = strings.TrimSpace(`
Ensure that your environment is up to to date with what is actually
configured in your dotfiles.

Goes through each of your manager's configuration files and running
ensure for each of them.`)

var ensureCmd = &cobra.Command{
	Use:   "ensure",
	Short: "Ensure your environment is up to date with your dotfiles",
	Long:  ensureLongMsg,
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
