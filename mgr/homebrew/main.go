package homebrew

import (
	"path/filepath"

	"github.com/mbark/punkt/exec"
)

// Dump will create a Brewfile using brew bundle, this file will be stored in
// the correct place in package structure and will be used to find what packages
// to install when calling ensure.
func Dump(dest string) {
	brewfile := filepath.Join(dest, "Brewfile")
	exec.Run("brew", "bundle", "dump", "--force", "--file="+brewfile)
}
