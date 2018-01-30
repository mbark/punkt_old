package symlink

import (
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/mbark/punkt/file"
	"github.com/mbark/punkt/path"
)

// Add ...
func (mgr Manager) Add(from, to string) error {
	symlink, err := mgr.addSymlink(from, to)
	if err != nil {
		return err
	}

	unexpanded := symlink.unexpend()
	mgr.saveSymlinks(unexpanded)

	return nil
}

func (mgr Manager) addSymlink(from, to string) (*Symlink, error) {
	if !filepath.IsAbs(from) {
		from = mgr.config.Fs.Join(mgr.config.WorkingDir, from)
	}

	pathFromHome, err := filepath.Rel(mgr.config.UserHome, from)
	if err != nil {
		logrus.WithError(err).Error("Unable to make target relative to user home")
		return nil, err
	}

	if to == "" {
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
		logrus.WithField("symlink", new).Warning("Unable to read file containing all symlinks, assuming non exists")
	}

	for _, existing := range saved {
		if new == existing {
			logrus.WithField("symlink", new).Info("Symlink already stored in file, not adding")
			return nil
		}
	}

	saved = append(saved, new)
	return file.SaveYaml(mgr.config.Fs, saved, mgr.config.Dotfiles, "symlinks")
}
