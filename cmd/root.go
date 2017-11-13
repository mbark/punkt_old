package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/kyokomi/emoji.v1"
)

var shortMessage = emoji.Sprint(":package: punkt; a dotfile manager to be dotty about")
var longMessage = emoji.Sprint(`:package: punkt manages your dotfiles and ensures that they match how your
environment actually looks. It can handle everything from simple
dotfile repos that just create a few symlinks, to those that
want to ensure all installed packages are kept up date.`)

// Opts contains the global run configuration options
var Opts = RunConfig{}

// RunConfig contains the configuration for running, primarily set by the
// command line arguments
type RunConfig struct {
	DryRun     bool
	LogLevel   string
	ConfigFile string
	ConfigDir  string
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "punkt",
	Short: shortMessage,
	Long:  longMessage,
	Run: func(cmd *cobra.Command, args []string) {
		setLogLevel(Opts.LogLevel)
	},
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
	cobra.OnInitialize(Opts.initConfig)
	RootCmd.PersistentFlags().StringVar(&Opts.ConfigFile, "config", "c", `Config file (default is $HOME/.punkt.yaml)`)
	RootCmd.PersistentFlags().StringVarP(&Opts.LogLevel, "log-level", "l", "info", `Set the logging level ("debug"|"info"|"warn"|"error"|"fatal")`)
	RootCmd.PersistentFlags().BoolVarP(&Opts.DryRun, "dry-run", "n", false, `Run through and print only`)
}

// initConfig reads in config file and ENV variables if set.
func (Opts *RunConfig) initConfig() {
	if Opts.ConfigFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(Opts.LogLevel)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".punkt" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".punkt")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// setLogLevel sets the logrus logging level
func setLogLevel(logLevel string) {
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
