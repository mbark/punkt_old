package git

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/mbark/punkt/run"
)

var fileRegexp = regexp.MustCompile(`file\:(?P<File>.*?)\s.*`)

// Dump ...
func Dump() []string {
	cmd := exec.Command("git", "config", "--list", "--show-origin", "--global")
	stdout := run.CaptureOutput(cmd)
	run.Run(cmd)

	output := strings.TrimSpace(stdout.String())
	rows := strings.Split(output, "\n")

	fileSet := make(map[string]struct{})

	for _, row := range rows {
		logrus.WithField("row", row).Info("Row")
		match := fileRegexp.FindStringSubmatch(row)
		logrus.WithField("match", match).Info("match")

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
