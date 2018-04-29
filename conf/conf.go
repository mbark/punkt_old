package conf

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/osfs"

	"github.com/mbark/punkt/file"
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
	Managers   map[string]map[string]string
}

// NewConfig builds a new configuration object from the given parameters
func NewConfig(configFile string) (*Config, error) {
	err := readConfigFile(configFile)
	if err != nil {
		return nil, err
	}

	setLogLevel()
	configureLogFiles()

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	fs := osfs.New("/")

	return &Config{
		PunktHome:  viper.GetString("punktHome"),
		Dotfiles:   viper.GetString("dotfiles"),
		UserHome:   path.GetUserHome(),
		WorkingDir: cwd,
		Fs:         fs,
		Command:    exec.Command,
		Managers:   readManagers(fs),
	}, nil
}

func readConfigFile(configFile string) error {
	abs, err := filepath.Abs(configFile)
	if err != nil {
		return errors.Wrapf(err, "failed to make config file path absolute: [file: %s]", configFile)
	}

	base := filepath.Base(abs)
	path := filepath.Dir(abs)
	fileName := strings.Split(base, ".")[0]

	viper.SetConfigName(fileName)
	viper.AddConfigPath(path)

	err = viper.ReadInConfig()
	return errors.Wrapf(err, "failed to read configuration [file: %s]", configFile)
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

func readManagers(fs billy.Filesystem) map[string]map[string]string {
	path := filepath.Join(viper.GetString("punktHome"), "managers.toml")

	var mgrs map[string]map[string]string
	err := file.ReadToml(fs, &mgrs, path)
	if err != nil {
		logrus.WithError(err).Error("Unable to read managers.toml")
	} else {
		logrus.WithFields(logrus.Fields{
			"managers": mgrs,
		}).Info("Found configuration for managers")
	}

	return mgrs
}
