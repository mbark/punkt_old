package symlink

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/mbark/punkt/file"
	"github.com/mbark/punkt/path"
)

// Add ...
func (mgr Manager) Add(from string) (*Symlink, error) {
	symlink, err := mgr.addSymlink(from)
	if err != nil {
		return nil, err
	}

	unexpanded := symlink.unexpend(mgr.config.UserHome)
	return symlink, mgr.saveSymlinks(*unexpanded)
}

func (mgr Manager) addSymlink(from string) (*Symlink, error) {
	if !filepath.IsAbs(from) {
		from = mgr.config.Fs.Join(mgr.config.WorkingDir, from)
	}

	pathFromHome, err := filepath.Rel(mgr.config.UserHome, from)
	if err != nil {
		logrus.WithError(err).Error("Unable to make target relative to user home")
		return nil, err
	}

	var to string
	if strings.HasPrefix(pathFromHome, "..") {
		to = mgr.config.Fs.Join(mgr.config.Dotfiles, from)
	} else {
		to = mgr.config.Fs.Join(mgr.config.Dotfiles, pathFromHome)
	}

	symlink := NewSymlink(mgr.config.Fs, to, from)
	logger := logrus.WithFields(logrus.Fields{
		"symlink": symlink,
	})

	if symlink.Exists() {
		logger.Info("Symlink already exists, not re-recreating")
		return symlink, nil
	}

	if _, err := mgr.config.Fs.Stat(to); err == nil {
		return nil, fmt.Errorf("File already exists where the file would be moved: %s", to)
	}

	err = path.CreateNecessaryDirectories(mgr.config.Fs, to)
	if err != nil {
		return nil, err
	}

	err = mgr.config.Fs.Rename(from, to)
	if err != nil {
		logger.WithError(err).Error("Unable to move target to destination placement")
		return nil, err
	}

	if err = symlink.Create(); err != nil {
		return nil, err
	}

	return symlink, nil
}

func (mgr Manager) saveSymlinks(new Symlink) error {
	saved := []Symlink{}
	// If we get an error reading the file, we ignore that and assume
	// that we can just continue and then overwrite the bad file
	err := file.Read(mgr.config.Fs, &saved, mgr.config.Dotfiles, "symlinks")
	if err != nil {
		logrus.WithError(err).WithField("symlink", new).Warning("Unable to read file containing all symlinks, assuming non exists")
	}

	for _, existing := range saved {
		if new.From == existing.From && new.To == existing.To {
			logrus.WithField("symlink", new).Info("Symlink already stored in file, not adding")
			return nil
		}
	}

	saved = append(saved, new)
	return file.SaveYaml(mgr.config.Fs, saved, mgr.config.Dotfiles, "symlinks")
}
