package cmd

import (
	"github.com/spf13/cobra"

	"github.com/mbark/punkt/brew"
	"github.com/mbark/punkt/git"
	"github.com/mbark/punkt/yarn"
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
	brew.Update()
	yarn.Update()
	git.Update(dotfiles, punktHome)
}
