package symlink

import (
	"os"
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
	from, err := filepath.Abs(from)
	if err != nil {
		return nil, err
	}

	relFrom, err := filepath.Rel(mgr.config.UserHome, from)
	if err != nil {
		return nil, err
	}

	if to == "" {
		to = filepath.Join(mgr.config.Dotfiles, relFrom)
	}

	symlink := NewSymlink(to, from)

	logger := logrus.WithFields(logrus.Fields{
		"symlink": symlink,
	})

	if symlink.Exists() {
		logger.Info("Symlink already exists, not re-recreating")
		return symlink, nil
	}

	err = path.CreateNecessaryDirectories(to)
	if err != nil {
		return nil, err
	}

	err = os.Rename(from, to)
	if err != nil {
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
	err := file.Read(&saved, mgr.config.Dotfiles, "symlinks")
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
	return file.SaveYaml(saved, mgr.config.Dotfiles, "symlinks")
}
