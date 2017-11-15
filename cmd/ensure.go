package cmd

import (
	"os"

	"github.com/mbark/punkt/backend"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// ensureCmd represents the ensure command
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
	hadError := false
	for _, symlink := range userConfig.Symlinks {
		hadError = !symlink.Create() || hadError
	}

	backend.CreatePackageDirectory(userConfig.PkgDbs)
	for name, backend := range userConfig.Backends {
		hadError = !backend.WriteInstalledPackages(name, userConfig.PkgDbs) || hadError
	}

	logrus.WithField("hadError", hadError).Info("All done")

	if hadError {
		os.Exit(1)
	}
}
