package brew

import (
	"os/exec"

	"github.com/mbark/punkt/run"
)

func bundle(args ...string) {
	arguments := append([]string{"bundle"}, args...)
	arguments = append(arguments, "--global")

	cmd := exec.Command("brew", arguments...)
	run.PrintOutputToUser(cmd)
	run.Run(cmd)
}
