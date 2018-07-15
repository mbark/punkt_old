package mgr

import (
	"path/filepath"
	"strings"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/mbark/punkt/pkg/conf"
	"github.com/mbark/punkt/pkg/fs"
	"github.com/mbark/punkt/pkg/mgr/generic"
	"github.com/mbark/punkt/pkg/mgr/git"
	"github.com/mbark/punkt/pkg/mgr/symlink"
	"github.com/mbark/punkt/pkg/printer"
)

// ManagerConfig ...
type ManagerConfig struct {
	Symlinks symlink.Config
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
	LinkManager symlink.LinkManager
	snapshot    fs.Snapshot
	config      conf.Config
}

// NewRootManager ...
func NewRootManager(config conf.Config, snapshot fs.Snapshot) *RootManager {
	return &RootManager{
		LinkManager: symlink.NewLinkManager(config, snapshot),
		snapshot:    snapshot,
		config:      config,
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

func (rootMgr RootManager) names(mgrs []Manager) string {
	var names []string
	for i := range mgrs {
		names = append(names, mgrs[i].Name())
	}

	return strings.Join(names, ", ")
}

// Dump ...
func (rootMgr RootManager) Dump(mgrs []Manager) error {
	printer.Log.Start("dump", "managers: <fg 2>%s", rootMgr.names(mgrs))

	var result error
	for i := range mgrs {
		printer.Log.Progress(i, len(mgrs), "running dump for <fg 2>%s manager", mgrs[i].Name())

		out, err := mgrs[i].Dump()
		if err != nil {
			printer.Log.Error("manager failed with error <fg 1>%s", err)
			result = multierror.Append(result, errors.Wrapf(err, "dump failed for %s", mgrs[i].Name()))
			continue
		}

		err = rootMgr.snapshot.Save(out, rootMgr.ConfigFile(mgrs[i].Name()))
		if err != nil {
			printer.Log.Error("failed to save configuration with error <fg 1>%s", err)
			result = multierror.Append(result, errors.Wrapf(err, "failed to save %s configuration", mgrs[i].Name()))
			continue
		}
	}

	printer.Log.Done("dump", "dump finished")
	return result
}

// Ensure ...
func (rootMgr RootManager) Ensure(mgrs []Manager) error {
	printer.Log.Start("ensure", "managers: <fg 2>%s", rootMgr.names(mgrs))

	var result error
	for i := range mgrs {
		printer.Log.Progress(i, len(mgrs), "running ensure for <fg 2>%s manager", mgrs[i].Name())

		logger := logrus.WithField("manager", mgrs[i].Name())
		logger.Debug("running ensure")

		err := mgrs[i].Ensure()
		if err != nil {
			printer.Log.Error("manager failed with error <fg 1>%s", err)
			result = multierror.Append(result, errors.Wrapf(err, "ensure failed for %s", mgrs[i].Name()))
			continue
		}

		config, err := rootMgr.readSymlinks(mgrs[i].Name())
		if err != nil {
			printer.Log.Error("failed to read stored symlinks with error <fg 1>%s", err)
			result = multierror.Append(result, errors.Wrapf(err, "unable to get %s configured symlinks", mgrs[i].Name()))
			continue
		}

		for i := range config.Symlinks {
			expanded := rootMgr.LinkManager.Expand(config.Symlinks[i])
			err = rootMgr.LinkManager.Ensure(expanded)
			if err != nil {
				printer.Log.Error("failed to create symlinks with error <fg 1>%s", err)
				result = multierror.Append(result, errors.Wrapf(err, "unable to ensure %s for manager %s", config.Symlinks[i], mgrs[i].Name()))
				continue
			}
		}
	}

	if result == nil {
		printer.Log.Done("ensure", "ensure finished")
	} else {
		printer.Log.Error("ensure did not successfully complete for all managers")
	}

	return result
}

// Update ...
func (rootMgr RootManager) Update(mgrs []Manager) error {
	printer.Log.Start("update", "managers: <fg 2>%s", rootMgr.names(mgrs))

	var result error
	for i := range mgrs {
		printer.Log.Progress(i, len(mgrs), "<fg 2>%s", mgrs[i].Name())

		err := mgrs[i].Update()
		if err != nil {
			printer.Log.Error("manager failed with error <fg 1>%s", err)
			result = multierror.Append(result, errors.Wrapf(err, "update failed for %s", mgrs[i].Name()))
			continue
		}
	}

	if result == nil {
		printer.Log.Done("update", "update finished")
	} else {
		printer.Log.Error("update did not successfully complete for all managers")
	}

	return result
}

func (rootMgr RootManager) readSymlinks(name string) (*symlink.Config, error) {
	var config ManagerConfig
	err := rootMgr.snapshot.ReadToml(&config, rootMgr.ConfigFile(name))
	if err != nil {
		if err == fs.ErrNoSuchFile {
			return &symlink.Config{Symlinks: []symlink.Symlink{}}, nil
		}

		return nil, err
	}

	return &config.Symlinks, nil
}

// Git ...
func (rootMgr RootManager) Git() git.Manager {
	return *git.NewManager(rootMgr.config, rootMgr.snapshot, rootMgr.ConfigFile("git"))
}

// Symlink ...
func (rootMgr RootManager) Symlink() symlink.Manager {
	return *symlink.NewManager(rootMgr.config, rootMgr.snapshot, rootMgr.ConfigFile("symlink"))
}

// ConfigFile ...
func (rootMgr RootManager) ConfigFile(name string) string {
	return filepath.Join(rootMgr.config.PunktHome, name+".toml")
}
