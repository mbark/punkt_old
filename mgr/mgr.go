package mgr

import (
	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/mgr/brew"
	"github.com/mbark/punkt/mgr/git"
	"github.com/mbark/punkt/mgr/symlink"
	"github.com/mbark/punkt/mgr/yarn"
)

// Manager ...
type Manager interface {
	Dump() error
	Ensure() error
	Update() error
}

// All returns a list of all available managers
func All(c conf.Config) []Manager {
	return []Manager{
		brew.NewManager(c),
		git.NewManager(c),
		symlink.NewManager(c),
		yarn.NewManager(c),
	}
}
