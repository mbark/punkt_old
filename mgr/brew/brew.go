package brew

import (
	"os/exec"
	"path/filepath"

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

func (mgr Manager) bundle(args ...string) {
	arguments := append([]string{"bundle"}, args...)
	arguments = append(arguments, "--global")

	cmd := exec.Command("brew", arguments...)
	run.PrintOutputToUser(cmd)
	run.Run(cmd)
}

// Dump ...
func (mgr Manager) Dump() {
	mgr.bundle("dump", "--force")
	brewfile := filepath.Join(path.GetUserHome(), ".Brewfile")
	symlinkMgr := symlink.NewManager(mgr.config)
	symlinkMgr.Add(brewfile, "")
}

// Ensure ...
func (mgr Manager) Ensure() {
	mgr.bundle("--no-upgrade")
}

// Update ...
func (mgr Manager) Update() {
	mgr.bundle()
}
