package git

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/mbark/punkt/run"
	"github.com/sirupsen/logrus"
)

var gitConfigFile = regexp.MustCompile(`file\:(?P<File>.*?)\s.*`)

func globalConfigFiles(command func(string, ...string) *exec.Cmd) ([]string, error) {
	// this is currently not suppported via the git library
	cmd := command("git", "config", "--list", "--show-origin", "--global")
	stdout, stderr := run.CaptureOutput(cmd)
	err := run.Run(cmd)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"stdout": stdout.String(),
			"stderr": stderr.String(),
		}).WithError(err).Error("Failed to run git config")
		return []string{}, err
	}

	logrus.WithFields(logrus.Fields{
		"stdout": stdout.String(),
	}).Debug("Got git config list successfully")

	output := strings.TrimSpace(stdout.String())
	rows := strings.Split(output, "\n")

	fileSet := make(map[string]struct{})

	for _, row := range rows {
		match := gitConfigFile.FindStringSubmatch(row)
		if len(match) > 1 {
			fileSet[match[1]] = struct{}{}
		}
	}

	files := []string{}
	for key := range fileSet {
		files = append(files, key)
	}

	return files, nil
}
