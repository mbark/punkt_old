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

// ManagerConfig ...
type ManagerConfig struct {
	Symlinks []symlink.Symlink
}

// Manager ...
type Manager interface {
	Name() string
	Dump() (string, error)
	Ensure() error
	Update() error
}

// RootManager ...
type RootManager struct {
	config conf.Config
}

// NewRootManager ...
func NewRootManager(config conf.Config) *RootManager {
	return &RootManager{
		config: config,
	}
}

// All returns a list of all available managers
func (rootMgr RootManager) All() []Manager {
	var mgrs []Manager
	for name := range rootMgr.config.Managers {
		mgr := generic.NewManager(rootMgr.config, rootMgr.ConfigFile(name), name)
		mgrs = append(mgrs, mgr)
	}

	return append(mgrs, rootMgr.Git(), rootMgr.Symlink())
}

// Dump ...
func (rootMgr RootManager) Dump(mgrs []Manager) error {
	for i := range mgrs {
		out, err := mgrs[i].Dump()
		if err != nil {
			return err
		}

		err = file.Save(rootMgr.config.Fs, out, rootMgr.ConfigFile(mgrs[i].Name()))
		if err != nil {
			return err
		}
	}

	return nil
}

// Ensure ...
func (rootMgr RootManager) Ensure(mgrs []Manager) error {
	for i := range mgrs {
		logger := logrus.WithField("manager", mgrs[i].Name())

		logger.Debug("running ensure")
		err := mgrs[i].Ensure()
		if err != nil {
			logger.WithError(err).Error("ensure failed")
			return err
		}

		symlinks, err := rootMgr.readSymlinks(mgrs[i].Name())
		if err != nil {
			logger.WithError(err).Error("unable to get symlinks")
			return err
		}

		linkManager := symlink.NewLinkManager(rootMgr.config)

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

// Update ...
func (rootMgr RootManager) Update(mgrs []Manager) error {
	for i := range mgrs {
		err := mgrs[i].Update()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"manager": mgrs[i],
			}).WithError(err).Error("Command ensure failed for manager")

			return err
		}
	}

	return nil
}

func (rootMgr RootManager) readSymlinks(name string) ([]symlink.Symlink, error) {
	var config ManagerConfig
	err := file.ReadToml(rootMgr.config.Fs, &config, rootMgr.ConfigFile(name))
	if err != nil && err != file.ErrNoSuchFile {
		if err == file.ErrNoSuchFile {
			return []symlink.Symlink{}, nil
		}

		return nil, err
	}

	return config.Symlinks, nil
}

// Git ...
func (rootMgr RootManager) Git() git.Manager {
	return *git.NewManager(rootMgr.config, rootMgr.ConfigFile("git"))
}

// Symlink ...
func (rootMgr RootManager) Symlink() symlink.Manager {
	return *symlink.NewManager(rootMgr.config, rootMgr.ConfigFile("symlink"))
}

// ConfigFile ...
func (rootMgr RootManager) ConfigFile(name string) string {
	return filepath.Join(rootMgr.config.PunktHome, name+".toml")
}
