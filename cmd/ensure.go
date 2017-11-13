package cmd

import (
	"os"

	"github.com/mbark/punkt/backends"
	"github.com/mbark/punkt/config"
	"github.com/mbark/punkt/symlinks"
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
		configFile := Opts.ConfigFile
		ensure(configFile)
	},
}

func init() {
	RootCmd.AddCommand(ensureCmd)
}

func ensure(configFile string) {
	conf := config.ParseConfig(configFile)
	logrus.WithFields(logrus.Fields{
		"config": conf,
	}).Debug("Successfully parsed config")

	hadError := symlinks.Create(conf, Opts.DryRun)
	backends.CreatePackageDirectory(conf, Opts.DryRun)

	for _, val := range conf.Backends {
		hadError = backends.WriteInstalledPackages(conf, val) && hadError
	}

	if hadError {
		os.Exit(1)
	}

}
