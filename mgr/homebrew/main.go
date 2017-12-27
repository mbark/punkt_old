package homebrew

import (
	"github.com/mbark/punkt/exec"
)

const brewfile = "~/.config/punkt/usr/Brewfile"

// Dump will create a Brewfile using brew bundle, this file will be stored in
// the correct place in package structure and will be used to find what packages
// to install when calling ensure.
func Dump() {
	exec.Run("brew", "bundle", "dump", "--force", "--file="+brewfile)
}
