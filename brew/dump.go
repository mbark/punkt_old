package brew

import (
	"path/filepath"

	"github.com/mbark/punkt/path"
)

// Dump ...
func Dump() string {
	bundle("dump", "--force")
	return filepath.Join(path.GetUserHome(), ".Brewfile")
}
