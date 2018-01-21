package symlink

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/mbark/punkt/path"
)

// Add ...
func Add(from, to, dotfiles string) *Symlink {
	from, err := filepath.Abs(from)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"from": from,
			"to":   to,
		}).WithError(err).Fatal("Unable to get absolute path")
	}

	relFrom, err := filepath.Rel(path.GetUserHome(), from)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"from": from,
			"to":   to,
		}).WithError(err).Fatal("Unable to get relative path")
	}

	if to == "" {
		to = filepath.Join(dotfiles, relFrom)
	}

	symlink := Symlink{
		From: to,
		To:   from,
	}

	logger := logrus.WithFields(logrus.Fields{
		"symlink": symlink,
	})

	if symlink.Exists() {
		logger.Info("Symlink already exists, not re-recreating")
		symlink = symlink.unexpend()
		return &symlink
	}

	path.CreateNecessaryDirectories(to)
	err = os.Rename(from, to)
	if err != nil {
		logger.WithError(err).Fatal("Unable to move file")
	}

	if symlink.Create() {
		symlink = symlink.unexpend()
		return &symlink
	}

	return nil
}
