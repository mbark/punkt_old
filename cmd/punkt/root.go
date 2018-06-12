package punkt

import (
	"fmt"
	"os"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/kyokomi/emoji.v1"

	"github.com/mbark/punkt/pkg/conf"
	"github.com/mbark/punkt/pkg/fs"
	"github.com/mbark/punkt/pkg/mgr"
)

var (
	logLevel   string
	configFile string
	punktHome  string
	dotfiles   string
)

var config *conf.Config
var snapshot *fs.Snapshot
var rootMgr mgr.RootManager

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "punkt",
	Short: emoji.Sprint(":package: punkt; a dotfile manager to be dotty about"),
	Long:  ``,
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
	RootCmd.Version = "0.0.1"

	var err error
	snapshot, err = fs.NewSnapshot()
	if err != nil {
		logrus.WithError(err).Fatal("failed to create filesystem snapshot")
		os.Exit(1)
	}

	configFile = snapshot.ExpandHome("~/.config/punkt/config.toml")
	punktHome = snapshot.ExpandHome("~/.config/punkt")
	dotfiles = snapshot.ExpandHome("~/.dotfiles")

	RootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", configFile, `The configuration file to read custom configuration from`)
	RootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", `Set the logging level ("debug"|"info"|"warn"|"error"|"fatal")`)
	RootCmd.PersistentFlags().StringVarP(&punktHome, "punkt-home", "p", punktHome, `Where all punkt configuration files should be stored`)
	RootCmd.PersistentFlags().StringVarP(&dotfiles, "dotfiles", "d", dotfiles, `The directory containing the user's dotfiles`)

	var result error
	err = viper.BindPFlag("logLevel", RootCmd.PersistentFlags().Lookup("log-level"))
	if err != nil {
		result = multierror.Append(result, err)
	}

	err = viper.BindPFlag("punktHome", RootCmd.PersistentFlags().Lookup("punkt-home"))
	if err != nil {
		result = multierror.Append(result, err)
	}

	err = viper.BindPFlag("dotfiles", RootCmd.PersistentFlags().Lookup("dotfiles"))
	if err != nil {
		result = multierror.Append(result, err)
	}

	if err != nil {
		logrus.WithError(result).Fatal("failed to bind flags to configuration")
	}
}

func initConfig() {
	var err error
	config, err = conf.NewConfig(*snapshot, snapshot.ExpandHome(configFile))
	if err != nil {
		logrus.WithError(err).Fatal("failed to red configuration file")
		os.Exit(1)
	}

	rootMgr = *mgr.NewRootManager(*config, *snapshot)
}
