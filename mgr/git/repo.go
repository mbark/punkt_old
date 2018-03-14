package git

import (
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	billy "gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-git.v4"
	gitconf "gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
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
		name = nameFromRemote(*config)
		if name == "" {
			name = filepath.Base(worktree.Root())
		}
	}

	return &Repo{
		Name:     name,
		Config:   *config,
		worktree: worktree,
	}, nil
}

func nameFromRemote(repo gitconf.Config) string {
	remote := repo.Remotes[git.DefaultRemoteName]
	if remote == nil {
		logrus.WithFields(logrus.Fields{
			"repo":    repo,
			"default": git.DefaultRemoteName,
		}).Debug("git repository doesn't have default remote")
		return ""

	}

	s := strings.Split(remote.URLs[0], "/")
	return s[len(s)-1]
}

// Exists ...
func (repo Repo) Exists() bool {
	logger := logrus.WithFields(logrus.Fields{
		"repo":     repo.Name,
		"worktree": repo.worktree,
	})
	logger.Debug("Checking if repo exists")

	s, err := filesystem.NewStorage(repo.worktree)
	if err != nil {
		logger.WithError(err).Error("Failed to create new storage for repository worktree")
		return false
	}

	r, err := git.Open(s, repo.worktree)

	if err == git.ErrRepositoryNotExists {
		return false
	}

	if r != nil {
		return true
	}

	return false
}

// Update ...
func (repo Repo) Update(reposDir string) error {
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
