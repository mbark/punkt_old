package brew

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

func bundle(args ...string) {
	arguments := append([]string{"bundle"}, args...)
	arguments = append(arguments, "--global")

	cmd := exec.Command("brew", arguments...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	logger := logrus.WithFields(logrus.Fields{
		"cmd": strings.Join(cmd.Args, " "),
	})

	logger.Info("Running command")
	err := cmd.Run()
	if err != nil {
		logger.WithFields(logrus.Fields{
			"stdout": stdout.String(),
			"stderr": stderr.String(),
		}).WithError(err).Fatal("Unable to run brew bundle command")
	}

	logger.WithFields(logrus.Fields{
		"stdout": stdout.String(),
		"stderr": stderr.String(),
	}).Debug("Command finished without error")
}
