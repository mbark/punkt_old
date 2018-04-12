package conf

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/osfs"

	"github.com/mbark/punkt/path"
)

// Config ...
type Config struct {
	PunktHome  string
	Dotfiles   string
	UserHome   string
	WorkingDir string
	Fs         billy.Filesystem
	Command    func(string, ...string) *exec.Cmd
}

// NewConfig builds a new configuration object from the given parameters
func NewConfig(configFile string) *Config {
	err := readConfigFile(configFile)
	if err != nil {
		os.Exit(1)
	}

	setLogLevel()
	configureLogFiles()

	cwd, err := os.Getwd()
	if err != nil {
		logrus.WithError(err).Fatal("Unable to determine working directory")
	}

	return &Config{
		PunktHome:  viper.GetString("punktHome"),
		Dotfiles:   viper.GetString("dotfiles"),
		UserHome:   path.GetUserHome(),
		WorkingDir: cwd,
		Fs:         osfs.New("/"),
		Command:    exec.Command,
	}
}

func readConfigFile(configFile string) error {
	abs, err := filepath.Abs(configFile)
	if err != nil {
		fmt.Printf("Failed to make %s an absolute path: %s\n", configFile, err)
		return err
	}

	base := filepath.Base(abs)
	path := filepath.Dir(abs)
	fileName := strings.Split(base, ".")[0]

	viper.SetConfigName(fileName)
	viper.AddConfigPath(path)

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Failed to read config file %s: %s\n", configFile, err)
		return err
	}

	return nil
}

func setLogLevel() {
	lvl, err := logrus.ParseLevel(viper.GetString("logLevel"))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"logLevel": lvl,
		}).WithError(err).Error("Unable to parse logging level")
		lvl = logrus.InfoLevel
	}

	logrus.WithField("level", lvl).Debug("Setting log level")
	logrus.SetLevel(lvl)
}

func configureLogFiles() {
	path := filepath.Join(viper.GetString("punktHome"), "punkt.log")
	writer, err := rotatelogs.New(
		path+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(path),
		rotatelogs.WithMaxAge(time.Duration(86400)*time.Second),
		rotatelogs.WithRotationTime(time.Duration(604800)*time.Second),
	)

	if err != nil {
		logrus.WithError(err).Error("Unable to create new writer")
	} else {
		logrus.SetOutput(writer)
	}
}
