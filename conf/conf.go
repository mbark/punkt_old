package conf

import (
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/mbark/punkt/path"
)

// Config ...
type Config struct {
	PunktHome string
	Dotfiles  string
	UserHome  string
}

// NewConfig builds a new configuration object from the given parameters
func NewConfig(configFile string) *Config {
	readConfigFile(configFile)
	setLogLevel()
	home := path.GetUserHome()

	return &Config{
		PunktHome: viper.GetString("punktHome"),
		Dotfiles:  viper.GetString("dotfiles"),
		UserHome:  home,
	}
}

func readConfigFile(configFile string) {
	logger := logrus.WithFields(logrus.Fields{
		"config": configFile,
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
