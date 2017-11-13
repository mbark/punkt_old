package symlinks

import (
	"os"
	"path/filepath"

	"github.com/mbark/punkt/config"
	"github.com/mbark/punkt/util"
	"github.com/sirupsen/logrus"
)

// Create will construct all symlinks specified in the configuration
// file. Returns true if all symlinks were successfully created, otherwise
// false. It will attempt to create all symlinks, even if one fails.
func Create(conf config.Config, dryRun bool) bool {
	hadError := false
	for to, from := range conf.Symlinks {
		successful := createOne(conf, from, to, dryRun)
		hadError = !successful || hadError
	}

	return hadError
}

func createOne(conf config.Config, from string, to string, dryRun bool) bool {
	from = filepath.Join(conf.ParentDir, from)
	to = filepath.Join(conf.ParentDir, to)

	logger := logrus.WithFields(logrus.Fields{
		"to":   to,
		"from": from,
	})

	_, err := os.Stat(from)
	if err != nil {
		logger.WithError(err).Warning("No such file")
		return false
	}

	if dryRun {
		return true
	}

	err = util.CreateNecessaryDirectories(to)
	if err != nil {
		logger.WithError(err).Warning("Unable to create necessary directories")
		return false
	}

	logger.Info("Creating symlink")
	err = os.Symlink(from, to)
	if err != nil {
		logrus.WithError(err).Warning("Unable to create symlink")
		return false
	}

	return true
}
