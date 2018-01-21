package yarn

import (
	"path/filepath"

	"github.com/mbark/punkt/path"
)

// Dump ...
func Dump() []string {
	configDir := filepath.Join(path.GetUserHome(), ".config", "yarn", "global")
	return []string{
		filepath.Join(configDir, "yarn.lock"),
		filepath.Join(configDir, "package.json"),
		filepath.Join(filepath.Join(path.GetUserHome(), ".yarnrc")),
	}
}
