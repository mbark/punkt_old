package git

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mbark/punkt/run"
	"github.com/sirupsen/logrus"
)

var fileRegexp = regexp.MustCompile(`file\:(?P<File>.*?)\s.*`)

func (mgr Manager) dumpConfig() ([]string, error) {
	// this is currently not suppported via the git library
	cmd := mgr.config.Command("git", "config", "--list", "--show-origin", "--global")
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
		match := fileRegexp.FindStringSubmatch(row)
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

func (mgr Manager) dumpRepos() ([]Repo, error) {
	repos := []Repo{}

	workingdir, err := mgr.config.Fs.Chroot(mgr.reposDir)
	if err != nil {
		return repos, nil
	}

	files, err := workingdir.ReadDir("./")
	logrus.WithFields(logrus.Fields{
		"files": files,
	}).Debug("Found the following files in the repos directory")

	if err != nil {
		logrus.WithError(err).Warning("Unable to read repos directory")
		return repos, err
	}

	for _, file := range files {
		if file.Mode()&os.ModeDir == 0 {
			continue
		}

		worktree, err := workingdir.Chroot(file.Name())
		if err != nil {
			return repos, nil
		}

		repo, err := NewRepo(worktree, filepath.Base(file.Name()))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"repo": worktree,
			}).WithError(err).Warning("Unable to open git repository")
			return repos, err
		}

		repos = append(repos, *repo)
	}

	logrus.WithFields(logrus.Fields{
		"repos": repos,
	}).Debug("Found git repos to save")
	return repos, nil
}
