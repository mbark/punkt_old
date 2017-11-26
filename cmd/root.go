package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/kyokomi/emoji.v1"
)

var (
	logLevel string
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
	RootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", `Set the logging level ("debug"|"info"|"warn"|"error"|"fatal")`)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	setLogLevel()
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
