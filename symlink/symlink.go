package symlink

import (
	"os"
	"path/filepath"

	"github.com/mbark/punkt/opt"
	"github.com/mbark/punkt/path"

	"github.com/sirupsen/logrus"
)

// Symlink contains information for a symlink to be created
type Symlink struct {
	To   string `yaml:"to"`
	From string `yaml:"from"`
}

// Create will construct the corresponding symlink. Returns true if the symlink
// was successfully created, otherwise false.
func (symlink Symlink) Create() bool {
	logger := logrus.WithFields(logrus.Fields{
		"to":   symlink.To,
		"from": symlink.From,
	})

	if symlink.exists() {
		logger.Info("Symlink already exists, not recreating")
		return true
	}

	_, err := os.Stat(symlink.From)
	if err != nil {
		logger.WithError(err).Error("No such file")
		return false
	}

	if opt.DryRun {
		return true
	}

	err = path.CreateNecessaryDirectories(symlink.To)
	if err != nil {
		logger.WithError(err).Error("Unable to create necessary directories")
		return false
	}

	logger.Info("Creating symlink")

	// It seems that os.symlink will do the symlink relative to the file that
	// we symlink to, meaning that from must be given either relative to
	// to the target or as an absolute path
	path, err := filepath.Abs(symlink.From)
	if err != nil {
		logrus.WithError(err).Error("Unable to convert path to absolute")
		return false
	}

	err = os.Symlink(path, symlink.To)
	if err != nil {
		logrus.WithError(err).Error("Unable to create symlink")
		return false
	}

	return true
}

func (symlink Symlink) exists() bool {
	from, _ := os.Stat(symlink.From)
	to, _ := os.Stat(symlink.To)
	logrus.WithFields(logrus.Fields{
		"to":   to,
		"from": from,
	}).Debug("Comparing if files are the same")
	return os.SameFile(from, to)
}
