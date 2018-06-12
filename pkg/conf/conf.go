package conf

import (
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/mbark/punkt/pkg/fs"
	"github.com/mbark/punkt/pkg/printer"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config ...
type Config struct {
	PunktHome string
	Dotfiles  string
	Managers  map[string]map[string]string
}

// NewConfig builds a new configuration object from the given parameters
func NewConfig(snapshot fs.Snapshot, file string) (*Config, error) {
	err := readConfig(snapshot, file)
	if err != nil {
		return nil, err
	}

	setLogLevel()
	configureLogFiles()

	mgrs, err := readManagers(snapshot)
	if err != nil {
		return nil, err
	}

	return &Config{
		PunktHome: viper.GetString("punktHome"),
		Dotfiles:  viper.GetString("dotfiles"),
		Managers:  mgrs,
	}, nil
}

func readConfig(snapshot fs.Snapshot, file string) error {
	abs, err := snapshot.AsAbsolute(file)
	if err != nil {
		return errors.Wrapf(err, "given config file %s does not exist", file)
	}

	printer.Log.Note("reading configuration from {fg 5}%s", snapshot.UnexpandHome(abs))

	f, err := snapshot.Fs.Open(abs)
	if err != nil {
		return errors.Wrapf(err, "failed to open %s", abs)
	}

	viper.SetConfigType("toml")
	err = viper.ReadConfig(f)
	if err != nil {
		return errors.Wrapf(err, "failed to read configuration from %s", abs)
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
		logrus.WithError(err).Error("unable to create rotatelogs writer")
	} else {
		logrus.SetOutput(writer)
	}
}

func readManagers(snapshot fs.Snapshot) (map[string]map[string]string, error) {
	path := filepath.Join(viper.GetString("punktHome"), "managers.toml")
	var mgrs map[string]map[string]string

	err := snapshot.ReadToml(&mgrs, path)
	if err != nil && err != fs.ErrNoSuchFile {
		return nil, errors.Wrapf(err, "unable to read manager configuration")
	}

	return mgrs, nil
}
