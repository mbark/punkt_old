package cmd

import (
	"fmt"
	"os"

	"github.com/mbark/punkt/backend"
	"github.com/mbark/punkt/opt"
	"github.com/mbark/punkt/symlink"

	"github.com/fatih/color"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/kyokomi/emoji.v1"
	"path/filepath"
)

var (
	configFile string
	logLevel   string
	dryRun     bool
)

var magenta = color.New(color.FgMagenta).SprintFunc()

var shortMessage = emoji.Sprint(":package: punkt; a dotfile manager to be dotty about")
var longMessage = emoji.Sprintf(`:package: %s manages your dotfiles and ensures that they match how your
environment actually looks. It can handle everything from simple
dotfile repos that just create a few symlinks, to those that
want to ensure all installed packages are kept up date.`, magenta("punkt"))

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "punkt",
	Short: shortMessage,
	Long:  longMessage,
}

var userConfig = UserConfig{}

// UserConfig is the parsed content of the user's configuration yaml file
type UserConfig struct {
	Symlinks []symlink.Symlink          `yaml:"symlinks"`
	Backends map[string]backend.Backend `yaml:"backends"`
	Tasks    []map[string]string        `yaml:"tasks"`
	PkgDbs   string                     `yaml:"pkgdbs"`
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&configFile, "config-file", "c", `Config file (default is $HOME/.punkt.yaml)`)
	RootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", `Set the logging level ("debug"|"info"|"warn"|"error"|"fatal")`)
	RootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false, `Run through and print only`)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		viper.AddConfigPath(home)
		viper.SetConfigName(".punkt")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		logrus.WithError(err).Fatal("Unable to find config file")
	}

	workingDir := filepath.Dir(viper.ConfigFileUsed())
	if err := os.Chdir(workingDir); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"workingdir": workingDir,
			"configFile": viper.ConfigFileUsed(),
		}).Fatal("Unable to change working directory to that of the config file")
	}

	setLogLevel()
	setUserConfig()
	opt.DryRun = dryRun
}

func setLogLevel() {
	if logLevel != "" {
		lvl, err := logrus.ParseLevel(logLevel)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to parse logging level: %s\n", logLevel)
			os.Exit(1)
		}
		logrus.SetLevel(lvl)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
}

// ReadUserConfig marshals the given config file to json
func setUserConfig() {
	err := viper.Unmarshal(&userConfig)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"file": viper.ConfigFileUsed(),
		}).WithError(err).Fatal("Unable to parse config file")
	}

	logrus.WithFields(logrus.Fields{
		"from":   viper.ConfigFileUsed(),
		"config": userConfig,
	}).Debug("Successfully parsed config")
}
