package git

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4/config"

	"github.com/mbark/punkt/run"
)

var fileRegexp = regexp.MustCompile(`file\:(?P<File>.*?)\s.*`)

// Dump ...
func Dump(punktHome string) ([]string, []config.Config) {
	return dumpConfig(), dumpRepos(punktHome)
}

func dumpConfig() []string {
	// this is currently not suppported via the git library
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

func dumpRepos(punktHome string) []config.Config {
	reposDir := reposDirectory(punktHome)
	logger := logrus.WithField("reposDir", reposDir)
	files, err := ioutil.ReadDir(reposDir)

	if err != nil {
		logger.WithError(err).Fatal("Unable to list files in the repos directory")
	}

	repos := []config.Config{}
	for _, file := range files {
		if file.Mode()&os.ModeDir == 0 {
			continue
		}

		repo := getRepo(filepath.Join(reposDir, file.Name()))
		if repo == nil {
			continue
		}

		conf, err := repo.Config()
		if err != nil {
			logger.WithError(err).Error("Unable to get git repo config")
		}

		repos = append(repos, *conf)
	}

	return repos
}
