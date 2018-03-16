package git

import (
	"fmt"
	"path/filepath"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/file"
	"github.com/mbark/punkt/mgr/symlink"
	billy "gopkg.in/src-d/go-billy.v4"

	"github.com/sirupsen/logrus"
)

// Manager ...
type Manager struct {
	config   conf.Config
	reposDir string
}

// NewManager ...
func NewManager(c conf.Config) *Manager {
	return &Manager{
		config:   c,
		reposDir: filepath.Join(c.PunktHome, "repos"),
	}
}

func (mgr Manager) repos() []Repo {
	repos := []Repo{}
	err := file.Read(mgr.config.Fs, &repos, mgr.config.Dotfiles, "repos")
	if err != nil {
		logrus.WithError(err).Warning("Unable to open repos.yml config file")
	}

	return repos
}

// Update ...
func (mgr Manager) Update() error {
	failed := []string{}
	for _, repo := range mgr.repos() {
		worktree, err := mgr.getWorktree(repo)
		if err != nil {
			failed = append(failed, repo.Name)
			continue
		}

		err = repo.Open(worktree)
		if err != nil {
			failed = append(failed, repo.Name)
			continue
		}

		err = repo.Update()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"repo": repo,
			}).WithError(err).Error("Unable to update git repository")
			failed = append(failed, repo.Name)
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("The following repos failed to update: %v", failed)
	}

	return nil
}

// Ensure ...
func (mgr Manager) Ensure() error {
	failed := []string{}

	repos := mgr.repos()
	logrus.WithField("repos", repos).Debug("Running ensure for these repos")

	for _, repo := range mgr.repos() {
		logger := logrus.WithField("repo", repo.Name)

		worktree, err := mgr.getWorktree(repo)
		if err != nil {
			failed = append(failed, repo.Name)
			continue
		}

		if err = repo.Open(worktree); err == nil {
			logrus.WithField("repo", repo).Debug("Repository already exists, skipping")
			continue
		}

		err = repo.Clone(worktree)
		if err != nil {
			logger.WithError(err).Error("Failed to clone repository")
			failed = append(failed, repo.Name)
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("The following repos failed to update: %v", failed)
	}

	return nil
}

func (mgr Manager) getWorktree(repo Repo) (billy.Filesystem, error) {
	worktree, err := mgr.config.Fs.Chroot(filepath.Join(mgr.reposDir, repo.Name))
	if err != nil {
		logrus.WithField("repo", repo.Name).WithError(err).Error("Failed to chroot to repo directory")
		return nil, err
	}

	return worktree, err

}

// Dump ...
func (mgr Manager) Dump() error {
	configFiles, err := mgr.dumpConfig()
	if err != nil {
		logrus.WithError(err).Error("Unable to find and save git configuration files")
		return err
	}

	symlinkMgr := symlink.NewManager(mgr.config)
	for _, f := range configFiles {
		err := symlinkMgr.Add(f, "")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"configFile": f,
			}).WithError(err).Warning("Unable to symlink git config file")
			return err
		}
	}

	repos, err := mgr.dumpRepos()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"reposDir": mgr.reposDir,
		}).WithError(err).Error("Unable to list repos")
		return err
	}

	return file.SaveYaml(mgr.config.Fs, repos, mgr.config.Dotfiles, "repos")
}
