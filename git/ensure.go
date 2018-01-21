package git

import (
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"

	"github.com/mbark/punkt/file"
)

// Ensure ...
func Ensure(dotfiles, punktHome string) {
	repos := []config.Config{}
	file.Read(&repos, dotfiles, "repos")

	reposDir := reposDirectory(punktHome)
	for _, repo := range repos {
		cloneDir := filepath.Join(reposDir, dirName(repo))
		if exists(repo, cloneDir) {
			logrus.WithFields(logrus.Fields{
				"dir":  cloneDir,
				"repo": repo,
			}).Debug("Repository already exists, skipping")
			continue
		}

		git.PlainClone(cloneDir, false, &git.CloneOptions{
			URL: repo.Remotes["origin"].URLs[0],
		})
	}
}

func exists(repo config.Config, path string) bool {
	r, err := git.PlainOpen(path)
	if err == git.ErrRepositoryNotExists {
		return false
	}

	if r != nil {
		return true
	}

	return false
}
