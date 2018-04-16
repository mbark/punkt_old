package cmd

import (
	"strings"

	"github.com/mbark/punkt/mgr"
	"github.com/spf13/cobra"
)

var message = strings.TrimSpace(`
dump the current working environment to your dotfile configuration files. Doing
this will search for symlinks and add your currently installed packages to their
corresponding package-manager's config files.`)

// ensureCmd represents the ensure command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "dump your current environment to config files",
	Long:  message,
	Run: func(cmd *cobra.Command, args []string) {
		dump(cmd, args)
	},
}

func init() {
	RootCmd.AddCommand(dumpCmd)
}

func dump(cmd *cobra.Command, args []string) {
	mgr.Dump(*config)
}
