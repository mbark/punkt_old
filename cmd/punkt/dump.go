package punkt

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var message = strings.TrimSpace(`
Dump the current working environment to your dotfile configuration.

Goes through all your specified managers and for each of these dumping
their configuration to their specific configuration files. This should
be free of side effects.`)

// ensureCmd represents the ensure command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump your current environment to your dotfiles directory",
	Long:  message,
	Run: func(cmd *cobra.Command, args []string) {
		dump(cmd, args)
	},
}

func init() {
	RootCmd.AddCommand(dumpCmd)
}

func dump(cmd *cobra.Command, args []string) {
	err := rootMgr.Dump(rootMgr.All())
	if err != nil {
		os.Exit(1)
	}
}
