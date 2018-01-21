package yarn

import (
	"os/exec"

	"github.com/mbark/punkt/run"
)

// Update ...
func Update() {
	cmd := exec.Command("yarn", "global", "upgrade")
	run.PrintOutputToUser(cmd)
	run.Run(cmd)
}
