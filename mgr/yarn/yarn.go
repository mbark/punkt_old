package yarn

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/mgr/symlink"
	"github.com/mbark/punkt/path"
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
	configDir := filepath.Join(path.GetUserHome(), ".config", "yarn", "global")
	symlinks := []string{
		filepath.Join(configDir, "yarn.lock"),
		filepath.Join(configDir, "package.json"),
		filepath.Join(filepath.Join(path.GetUserHome(), ".yarnrc")),
	}

	for _, s := range symlinks {
		symlinkMgr := symlink.NewManager(mgr.config)
		symlinkMgr.Add(s, "")
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
