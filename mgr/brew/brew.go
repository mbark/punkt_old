package brew

import (
	"path/filepath"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/mgr/symlink"
	"github.com/mbark/punkt/run"
	"github.com/sirupsen/logrus"
)

// Manager ...
type Manager struct {
	config conf.Config
}

// NewManager ...
func NewManager(c conf.Config) *Manager {
	return &Manager{
		config: c,
	}
}

func (mgr Manager) bundle(args ...string) error {
	arguments := append([]string{"bundle"}, args...)
	arguments = append(arguments, "--global")

	cmd := mgr.config.Command("brew", arguments...)
	run.PrintOutputToUser(cmd)
	return run.Run(cmd)
}

// Dump ...
func (mgr Manager) Dump() error {
	err := mgr.bundle("dump", "--force")
	if err != nil {
		logrus.WithError(err).Error("Unable to run dump bundle")
		return err
	}

	brewfile := filepath.Join(mgr.config.UserHome, ".Brewfile")
	symlinkMgr := symlink.NewManager(mgr.config)
	_, err = symlinkMgr.Add(brewfile)
	return err
}

// Ensure ...
func (mgr Manager) Ensure() error {
	return mgr.bundle("--no-upgrade")
}

// Update ...
func (mgr Manager) Update() error {
	return mgr.bundle()
}
