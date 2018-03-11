package git

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	billy "gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-git.v4"
	gitconf "gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

// Repo describes a git repository
type Repo struct {
	Name     string
	Config   gitconf.Config
	worktree billy.Filesystem
}

// NewRepo opens the repository at the given worktree with a filesytem storage as
// backup.
func NewRepo(worktree billy.Filesystem, name string) (*Repo, error) {
	s, err := filesystem.NewStorage(worktree)
	if err != nil {
		return nil, err
	}

	repo, err := git.Open(s, worktree)
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

	return &Repo{
		Name:     name,
		Config:   *config,
		worktree: worktree,
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

func (repo Repo) exists() bool {
	r, err := git.Open(memory.NewStorage(), repo.worktree)
	if err == git.ErrRepositoryNotExists {
		return false
	}

	if r != nil {
		return true
	}

	return false
}

func (repo Repo) update(reposDir string) error {
	// TODO: change to Open with storer and billy.Filesystem
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
