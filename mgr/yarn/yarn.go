package yarn

import (
	"os/exec"
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

// Dump ...
func (mgr Manager) Dump() {
	configDir := filepath.Join(mgr.config.UserHome, ".config", "yarn", "global")
	symlinks := []string{
		filepath.Join(configDir, "yarn.lock"),
		filepath.Join(configDir, "package.json"),
		filepath.Join(filepath.Join(mgr.config.UserHome, ".yarnrc")),
	}

	for _, s := range symlinks {
		symlinkMgr := symlink.NewManager(mgr.config)
		err := symlinkMgr.Add(s, "")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"symlink": s,
			}).WithError(err).Fatal("Unable to add symlink")
		}
	}
}

// Ensure ...
func (mgr Manager) Ensure() {
	cmd := exec.Command("yarn")
	run.PrintOutputToUser(cmd)
	cmd.Dir = workingDir()

	run.Run(cmd)
}

func workingDir() string {
	cmd := exec.Command("yarn", "global", "dir")
	stdout := run.CaptureOutput(cmd)
	run.Run(cmd)

	return strings.TrimSpace(stdout.String())
}

// Update ...
func (mgr Manager) Update() {
	cmd := exec.Command("yarn", "global", "upgrade")
	run.PrintOutputToUser(cmd)
	run.Run(cmd)
}
