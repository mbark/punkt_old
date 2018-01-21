package yarn

import (
	"os/exec"
	"strings"

	"github.com/mbark/punkt/run"
)

// Ensure ...
func Ensure() {
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
