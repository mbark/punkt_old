package cmd

import (
	"github.com/spf13/cobra"

	"github.com/mbark/punkt/brew"
	"github.com/mbark/punkt/file"
	"github.com/mbark/punkt/symlink"
	"github.com/mbark/punkt/yarn"
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
	symlinks := []symlink.Symlink{}
	file.Read(&symlinks, dotfiles, "symlinks")
	symlink.Ensure(symlinks)

	brew.Ensure()
	yarn.Ensure()
}
