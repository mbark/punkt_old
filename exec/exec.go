package exec

import (
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
	"strings"
)

// Run runs the given command with args, logging stdout and stderr if the program
// errors out.
func Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	logger := logrus.WithFields(logrus.Fields{
		"cmd": strings.Join(cmd.Args, " "),
	})

	logger.Info("Running command")
	err := cmd.Run()
	if err != nil {
		logger.WithError(err).Error("Unable to create Brewfile")
	}

	return err
}
