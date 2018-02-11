package git

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mbark/punkt/run"
)

var fileRegexp = regexp.MustCompile(`file\:(?P<File>.*?)\s.*`)

func dumpConfig() ([]string, error) {
	// this is currently not suppported via the git library
	cmd := exec.Command("git", "config", "--list", "--show-origin", "--global")
	stdout, _ := run.CaptureOutput(cmd)
	err := run.Run(cmd)
	if err != nil {
		return []string{}, err
	}

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

func dumpRepos(reposDir string) ([]gitRepo, error) {
	files, err := ioutil.ReadDir(reposDir)
	repos := []gitRepo{}

	if err != nil {
		return repos, err
	}

	for _, file := range files {
		if file.Mode()&os.ModeDir == 0 {
			continue
		}

		repo, err := newGitRepo(filepath.Join(reposDir, file.Name()), "")
		if err != nil {
			return repos, err
		}

		repos = append(repos, *repo)
	}

	return repos, nil
}
