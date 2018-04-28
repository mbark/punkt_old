package mgr

import (
	"path/filepath"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/file"
	"github.com/mbark/punkt/mgr/generic"
	"github.com/mbark/punkt/mgr/git"
	"github.com/mbark/punkt/mgr/symlink"
	"github.com/sirupsen/logrus"
)

// Manager ...
type Manager interface {
	Name() string
	Dump() (string, error)
	Ensure() error
	Update() error
	Symlinks() ([]symlink.Symlink, error)
}

// All returns a list of all available managers
func All(c conf.Config) []Manager {
	var mgrs []Manager
	for name := range c.Managers {
		mgr := generic.NewManager(name, configFile(c, name), c)
		mgrs = append(mgrs, mgr)
	}

	return append(mgrs, Git(c), Symlink(c))
}

// Dump ...
func Dump(c conf.Config) error {
	mgrs := All(c)
	for i := range mgrs {
		out, err := mgrs[i].Dump()
		if err != nil {
			return err
		}

		err = file.Save(c.Fs, out, configFile(c, mgrs[i].Name()))
		if err != nil {
			return err
		}
	}

	return nil
}

// Ensure ...
func Ensure(c conf.Config) error {
	mgrs := All(c)
	for i := range mgrs {
		logger := logrus.WithField("manager", mgrs[i].Name())

		logger.Debug("running ensure")
		err := mgrs[i].Ensure()
		if err != nil {
			logger.WithError(err).Error("ensure failed")
			return err
		}

		symlinks, err := mgrs[i].Symlinks()
		if err != nil {
			logger.WithError(err).Error("unable to get symlinks")
			return err
		}

		linkManager := symlink.NewLinkManager(c)

		logger = logger.WithField("symlinks", symlinks)
		for i := range symlinks {
			expanded := linkManager.Expand(symlinks[i])
			err = linkManager.Ensure(expanded)
			if err != nil {
				logger.WithField("symlink", symlinks[i]).WithError(err).Error("unable to ensure symlink")
				return err
			}
		}
	}

	return nil
}

// Git ...
func Git(c conf.Config) git.Manager {
	return *git.NewManager(c, configFile(c, "git"))
}

// Symlink ...
func Symlink(c conf.Config) symlink.Manager {
	return *symlink.NewManager(c, configFile(c, "symlink"))
}

func configFile(c conf.Config, name string) string {
	return filepath.Join(c.PunktHome, name+".toml")
}
