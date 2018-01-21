package git

import (
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
)

// Update ...
func Update(dotfiles, punktHome string) {
	repos := getRepos(dotfiles)
	reposDir := reposDirectory(punktHome)
	for _, repo := range repos {
		cloneDir := filepath.Join(reposDir, dirName(repo))

		logger := logrus.WithFields(logrus.Fields{
			"repo": repo,
			"dir":  cloneDir,
		})

		r, err := git.PlainOpen(cloneDir)
		if err != nil {
			logger.WithError(err).Fatal("Unable to open git repository")
		}

		w, err := r.Worktree()
		if err != nil {
			logger.WithError(err).Fatal("Unable to get working tree of git repository")
		}

		err = w.Pull(&git.PullOptions{RemoteName: "origin"})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			logger.WithError(err).Fatal("Unable to pull git repository")
		}
	}
}
