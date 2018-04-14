package mgr

import (
	"path/filepath"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/mgr/generic"
)

// Manager ...
type Manager interface {
	Dump() (string, error)
	Ensure() error
	Update() (string, error)
}

// All returns a list of all available managers
func All(c conf.Config) []Manager {
	var mgrs []Manager
	for name := range c.Managers {
		mgr := generic.NewManager(name, configFile(c, name), c)
		mgrs = append(mgrs, mgr)
	}

	return mgrs
}

func configFile(c conf.Config, name string) string {
	return filepath.Join(c.PunktHome, name+".toml")
}
