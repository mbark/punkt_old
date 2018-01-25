package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	configFile = path.ExpandHome("~/.config/punkt/config")
	punktHome  = path.ExpandHome("~/.config/punkt")
	dotfiles   = path.ExpandHome("~/.dotfiles")
)

var config *conf.Config
var managers []mgr.Manager

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

	RootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", configFile, `The configuration file to read custom configuration from`)
	RootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", `Set the logging level ("debug"|"info"|"warn"|"error"|"fatal")`)
	RootCmd.PersistentFlags().StringVarP(&punktHome, "punkt-home", "p", punktHome, `Where all punkt configuration files should be stored`)
	RootCmd.PersistentFlags().StringVarP(&dotfiles, "dotfiles", "d", dotfiles, `The directory containing the user's dotfiles`)

	viper.BindPFlag("logLevel", RootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("punktHome", RootCmd.PersistentFlags().Lookup("punkt-home"))
	viper.BindPFlag("dotfiles", RootCmd.PersistentFlags().Lookup("dotfiles"))
}

func initConfig() {
	readConfigFile()
	setLogLevel()

	home := path.GetUserHome()
	config = conf.NewConfig(punktHome, dotfiles, home)
	managers = mgr.All(*config)
}

func readConfigFile() {
	logger := logrus.WithFields(logrus.Fields{
		"config": config,
	})

	configFile = path.ExpandHome(configFile)
	logger.Info("Reading configuration file")

	abs, err := filepath.Abs(configFile)
	if err != nil {
		logger.WithError(err).Error("Error reading provided configuration file")
	}

	base := filepath.Base(abs)
	path := filepath.Dir(abs)

	viper.SetConfigName(strings.Split(base, ".")[0])
	viper.AddConfigPath(path)

	if err := viper.ReadInConfig(); err != nil {
		logger.WithError(err).Error("Failed to read config file")
	}
}

func setLogLevel() {
	lvl, err := logrus.ParseLevel(viper.GetString("logLevel"))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"level": lvl,
		}).WithError(err).Fatal("Unable to parse logging level")
	}

	logrus.SetLevel(lvl)
}
