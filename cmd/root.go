package cmd

import (
	"fmt"
	"os"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/kyokomi/emoji.v1"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/mgr"
	"github.com/mbark/punkt/path"
)

var (
	logLevel   string
	configFile = path.ExpandHome("~/.config/punkt/config", path.GetUserHome())
	punktHome  = path.ExpandHome("~/.config/punkt", path.GetUserHome())
	dotfiles   = path.ExpandHome("~/.dotfiles", path.GetUserHome())
)

var config *conf.Config
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

	RootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", configFile, `The configuration file to read custom configuration from`)
	RootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", `Set the logging level ("debug"|"info"|"warn"|"error"|"fatal")`)
	RootCmd.PersistentFlags().StringVarP(&punktHome, "punkt-home", "p", punktHome, `Where all punkt configuration files should be stored`)
	RootCmd.PersistentFlags().StringVarP(&dotfiles, "dotfiles", "d", dotfiles, `The directory containing the user's dotfiles`)

	var result error
	err := viper.BindPFlag("logLevel", RootCmd.PersistentFlags().Lookup("log-level"))
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
	config, err = conf.NewConfig(path.ExpandHome(configFile, path.GetUserHome()))
	if err != nil {
		logrus.WithError(err).Fatal("failed to red configuration file")
		os.Exit(1)
	}

	rootMgr = *mgr.NewRootManager(*config)
}
