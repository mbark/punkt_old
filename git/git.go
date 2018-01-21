package git

import (
	"path/filepath"
	"strings"

	"github.com/mbark/punkt/file"

	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
)

func getRepos(dotfiles string) []config.Config {
	repos := []config.Config{}
	file.Read(&repos, dotfiles, "repos")

	return repos
}

func reposDirectory(punktHome string) string {
	return filepath.Join(punktHome, "repos")
}

func dirName(repo config.Config) string {
	s := strings.Split(repo.Remotes["origin"].URLs[0], "/")
	return s[len(s)-1]
}

func getRepo(cloneDir string) *git.Repository {
	logger := logrus.WithFields(logrus.Fields{
		"dir": cloneDir,
	})

	r, err := git.PlainOpen(cloneDir)
	if err != nil {
		logger.WithError(err).Error("Unable to open git repository")
		return nil
	}

	return r
}

func getWorktree(repo *git.Repository) *git.Worktree {
	w, err := repo.Worktree()
	if err != nil {
		logrus.WithField("repo", repo).WithError(err).Error("Unable to get working tree of git repository")
		return nil
	}

	return w
}
