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
	Name       string
	Config     *gitconf.Config
	repository *git.Repository
}

// NewRepo opens the repository at the given worktree with a filesytem storage as
// a backing storage.
func NewRepo(name string) *Repo {
	return &Repo{
		Name:       name,
		Config:     nil,
		repository: nil,
	}
}

// OpenRepo is a convience method to create a new repo and open it.
func OpenRepo(worktree billy.Filesystem, name string) (*Repo, error) {
	repo := NewRepo(name)
	err := repo.Open(worktree)

	return repo, err
}

// Open ...
func (repo *Repo) Open(worktree billy.Filesystem) error {
	s, err := filesystem.NewStorage(worktree)
	if err != nil {
		return err
	}

	repository, err := git.Open(s, worktree)
	if err != nil {
		return err
	}

	config, err := repository.Config()
	if err != nil {
		return err
	}

	name := repo.Name
	if name == "" {
		name = nameFromRemote(*config)
		if name == "" {
			name = filepath.Base(worktree.Root())
		}
	}

	repo.Name = name
	repo.Config = config
	repo.repository = repository
	return nil
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

// Clone ...
func (repo *Repo) Clone(worktree billy.Filesystem) error {
	storer, err := filesystem.NewStorage(worktree)
	if err != nil {
		logrus.WithField("repo", repo.Name).WithError(err).Error("Unable to create storage for repo")
		return err
	}

	remote := repo.Config.Remotes["origin"].URLs[0]
	logrus.WithFields(logrus.Fields{
		"repo":   repo.Name,
		"remote": remote,
	}).Debug("Cloning repository from remote")

	repository, err := git.Clone(storer, worktree, &git.CloneOptions{
		URL: remote,
	})

	repo.repository = repository
	return err
}

// Update ...
func (repo Repo) Update() error {
	w, err := repo.repository.Worktree()
	if err != nil {
		return err
	}

	err = w.Pull(&git.PullOptions{RemoteName: git.DefaultRemoteName})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	return nil
}
