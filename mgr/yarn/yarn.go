package yarn

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/mgr/symlink"
	"github.com/mbark/punkt/run"
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

// ConfigFiles ...
func (mgr Manager) ConfigFiles() []string {
	configDir := filepath.Join(mgr.config.UserHome, ".config", "yarn", "global")
	return []string{
		filepath.Join(configDir, "yarn.lock"),
		filepath.Join(configDir, "package.json"),
		filepath.Join(filepath.Join(mgr.config.UserHome, ".yarnrc")),
	}
}

// Dump ...
func (mgr Manager) Dump() error {
	failed := []string{}
	for _, s := range mgr.ConfigFiles() {
		symlinkMgr := symlink.NewManager(mgr.config)
		err := symlinkMgr.Add(s, "")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"symlink": s,
			}).WithError(err).Error("Unable to add symlink")
			failed = append(failed, s)
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("The following symlinks could not be added: %v", failed)
	}

	return nil
}

// Ensure ...
func (mgr Manager) Ensure() error {
	dir, err := mgr.globalDir()
	if err != nil {
		return err
	}

	logrus.WithField("globalDir", dir).Debug("Yarn global dir")

	cmd := mgr.config.Command("yarn")
	cmd.Dir = dir
	run.PrintOutputToUser(cmd)
	return run.Run(cmd)
}

func (mgr Manager) globalDir() (string, error) {
	cmd := mgr.config.Command("yarn", "global", "dir")
	stdout, stderr := run.CaptureOutput(cmd)
	err := run.Run(cmd)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"stdout": stdout.String(),
			"stderr": stderr.String(),
		}).Error("Unable to determine global directory for yarn")
		return "", err
	}

	return strings.TrimSpace(stdout.String()), nil
}

// Update ...
func (mgr Manager) Update() error {
	cmd := mgr.config.Command("yarn", "global", "upgrade")
	run.PrintOutputToUser(cmd)
	return run.Run(cmd)
}
