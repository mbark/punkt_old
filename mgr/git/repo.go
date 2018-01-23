package git

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
	gitconf "gopkg.in/src-d/go-git.v4/config"
)

// GitRepo ...
type gitRepo struct {
	Name   string
	Config gitconf.Config
	path   string
}

func newGitRepo(path, name string) (*gitRepo, error) {
	repo, err := git.PlainOpen(path)
	if repo == nil {
		return nil, err
	}

	config, err := repo.Config()
	if err != nil {
		return nil, err
	}

	if name == "" {
		name, err = repoName(*config)
		if err != nil {
			return nil, err
		}
	}

	return &gitRepo{
		Name:   name,
		Config: *config,
		path:   path,
	}, nil
}

func repoName(repo gitconf.Config) (string, error) {
	remote := repo.Remotes[git.DefaultRemoteName]
	if remote == nil {
		logrus.WithFields(logrus.Fields{
			"repo":    repo,
			"default": git.DefaultRemoteName,
		}).Warning("git repository doesn't have default remote name")

		if len(repo.Remotes) == 0 {
			return "", errors.New("git repository has no remotes")
		}

		for _, r := range repo.Remotes {
			remote = r
			break
		}

	}

	s := strings.Split(remote.URLs[0], "/")
	return s[len(s)-1], nil
}

func (repo gitRepo) exists() bool {
	r, err := git.PlainOpen(repo.path)
	if err == git.ErrRepositoryNotExists {
		return false
	}

	if r != nil {
		return true
	}

	return false
}

func (repo gitRepo) update(reposDir string) error {
	r, err := git.PlainOpen(filepath.Join(reposDir, repo.Name))
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	err = w.Pull(&git.PullOptions{RemoteName: git.DefaultRemoteName})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	return nil
}
