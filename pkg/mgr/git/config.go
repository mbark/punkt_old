package git

import (
	"regexp"
	"strings"

	"github.com/mbark/punkt/pkg/run"
	"github.com/sirupsen/logrus"
)

var gitConfigFile = regexp.MustCompile(`file\:(?P<File>.*?)\s.*`)

func globalConfigFiles() []string {
	// this is currently not suppported via the git library
	cmd := run.Commander("git", "config", "--list", "--show-origin", "--global")
	out, err := cmd.Output()

	stdout := string(out)
	logger := logrus.WithFields(logrus.Fields{
		"stdout": stdout,
	})

	if err != nil {
		logger.WithError(err).Error("Failed to find git config files")
		return []string{}
	}

	logger.Debug("Got git config list successfully")

	output := strings.TrimSpace(stdout)
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

	return files
}
