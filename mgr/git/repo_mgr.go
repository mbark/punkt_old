package git

import (
	"path/filepath"

	"github.com/sirupsen/logrus"
	billy "gopkg.in/src-d/go-billy.v4"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
)

// RepoManager ...
type RepoManager interface {
	Dump(dir string) (*Repo, error)
	Ensure(repo Repo) error
	Update(dir string) (bool, error)
}

// GoGitRepoManager ...
type goGitRepoManager struct {
	fs billy.Filesystem
}

// NewRepoManager ...
func NewRepoManager(fs billy.Filesystem) RepoManager {
	return goGitRepoManager{
		fs: fs,
	}
}

func (mgr goGitRepoManager) storage(dir string) (storage.Storer, billy.Filesystem, error) {
	logrus.WithField("dir", dir).Debug("Constructing storage for directory")
	worktree, err := mgr.fs.Chroot(dir)
	if err != nil {
		return nil, nil, err
	}

	dotGit, err := worktree.Chroot(".git")
	if err != nil {
		return nil, nil, err
	}

	storage, err := filesystem.NewStorage(dotGit)
	if err != nil {
		return nil, nil, err
	}

	return storage, worktree, nil
}

func (mgr goGitRepoManager) open(dir string) (*git.Repository, error) {
	storage, worktree, err := mgr.storage(dir)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"directory": dir,
		}).WithError(err).Debug("Failed to create storage for repository")
		return nil, err
	}

	return git.Open(storage, worktree)
}

// Dump ...
func (mgr goGitRepoManager) Dump(dir string) (*Repo, error) {
	repository, err := mgr.open(dir)
	if err != nil {
		return nil, err
	}

	config, err := repository.Config()
	if err != nil {
		return nil, err
	}

	dir, err = filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	return &Repo{
		Name:   filepath.Base(dir),
		Config: config,
		Path:   dir,
	}, nil
}

// Ensure ...
func (mgr goGitRepoManager) Ensure(repo Repo) error {
	logger := logrus.WithFields(logrus.Fields{
		"repo": repo.Name,
		"path": repo.Path,
	})
	logger.Info("Ensuring repository exists")

	if _, ok := mgr.open(repo.Path); ok == nil {
		logger.Info("Repository already exists")
		return nil
	}

	storage, worktree, err := mgr.storage(repo.Path)
	if err != nil {
		logger.WithError(err).Error("unable to allocate storage for repository")
		return err
	}

	remote := repo.Config.Remotes[git.DefaultRemoteName].URLs[0]

	logger = logger.WithField("remote", remote)
	logger.Debug("Cloning repository from remote")
	repository, err := git.Clone(storage, worktree, &git.CloneOptions{
		URL: remote,
	})
	if err != nil {
		logger.WithError(err).Error("unable to clone repository")
		return err
	}

	err = repository.Storer.SetConfig(repo.Config)
	if err != nil {
		logger.WithError(err).Error("unable to set configuration for repository")
		return err
	}

	return nil
}

// Update ...
func (mgr goGitRepoManager) Update(dir string) (bool, error) {
	logger := logrus.WithField("repo", dir)
	logger.Info("Updating repository")

	repository, err := mgr.open(dir)
	if err != nil {
		logger.WithError(err).Error("unable to open git repository")
		return false, err
	}

	w, err := repository.Worktree()
	if err != nil {
		logger.WithError(err).Error("unable to get worktree for repository")
		return false, err
	}

	updated := true
	err = w.Pull(&git.PullOptions{RemoteName: git.DefaultRemoteName})
	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			logger.Info("repository is already up to date")
			updated = false
		} else {
			logger.WithError(err).Error("unable to update repository")
			return false, err
		}
	}

	logger.Info("Repository successfully updated")
	return updated, nil
}
