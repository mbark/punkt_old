package symlink

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/mbark/punkt/file"
	"github.com/mbark/punkt/path"
)

// Add ...
func (mgr Manager) Add(from, to string) {
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
		to = filepath.Join(mgr.config.Dotfiles, relFrom)
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
		return
	}

	path.CreateNecessaryDirectories(to)
	err = os.Rename(from, to)
	if err != nil {
		logger.WithError(err).Fatal("Unable to move file")
	}

	if !symlink.Create() {
		return
	}

	symlink = symlink.unexpend()
	mgr.saveSymlink(symlink)
}

func (mgr Manager) saveSymlink(symlink Symlink) {
	symlinks := []Symlink{}
	file.Read(&symlinks, mgr.config.Dotfiles, "symlinks")

	for _, s := range symlinks {
		if s == symlink {
			logrus.WithField("symlink", symlink).Info("Symlink already stored in file, not adding")
			return
		}
	}

	symlinks = append(symlinks, symlink)

	file.SaveYaml(symlinks, mgr.config.Dotfiles, "symlinks")
}
