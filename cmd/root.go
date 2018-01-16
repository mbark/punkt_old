package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mbark/punkt/path"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/kyokomi/emoji.v1"
)

var (
	defaultConfig    = path.ExpandHome("~/.config/punkt/config")
	defaultPunktHome = path.ExpandHome("~/.config/punkt")
	defaultDotfiles  = path.ExpandHome("~/.dotfiles")
	logLevel         string
	config           string
	punktHome        string
	dotfiles         string
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

	RootCmd.PersistentFlags().StringVarP(&config, "config", "c", defaultConfig, `The configuration file to read custom configuration from`)
	RootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", `Set the logging level ("debug"|"info"|"warn"|"error"|"fatal")`)
	RootCmd.PersistentFlags().StringVarP(&punktHome, "punkt-home", "p", defaultPunktHome, `Where all punkt configuration files should be stored`)
	RootCmd.PersistentFlags().StringVarP(&dotfiles, "dotfiles", "d", defaultDotfiles, `The directory containing the user's dotfiles`)

	viper.BindPFlag("logLevel", RootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("punktHome", RootCmd.PersistentFlags().Lookup("punkt-home"))
	viper.BindPFlag("dotfiles", RootCmd.PersistentFlags().Lookup("dotfiles"))
}

func initConfig() {
	readConfigFile()
	setLogLevel()
}

func readConfigFile() {
	if config == "" {
		config = defaultConfig
	}

	logger := logrus.WithFields(logrus.Fields{
		"config": config,
	})

	config = path.ExpandHome(config)
	logger.Info("Reading configuration file")

	abs, err := filepath.Abs(config)
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
	lvl, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"level": lvl,
		}).WithError(err).Fatal("Unable to parse logging level")
	}

	logrus.SetLevel(lvl)
}
