package git

import (
	"fmt"
	"path/filepath"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/file"
	"github.com/mbark/punkt/mgr/symlink"

	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
)

// Manager ...
type Manager struct {
	config conf.Config
}

// NewManager ...
func NewManager(c conf.Config) *Manager {
	return &Manager{
		config: c,
	}
}

func (mgr Manager) reposDir() string {
	return filepath.Join(mgr.config.PunktHome, "repos")
}

func (mgr Manager) repos() []gitRepo {
	repos := []gitRepo{}
	file.Read(mgr.config.Fs, &repos, mgr.config.Dotfiles, "repos")

	return repos
}

// Update ...
func (mgr Manager) Update() error {
	failed := []string{}
	for _, gitRepo := range mgr.repos() {
		err := gitRepo.update(mgr.reposDir())
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"gitRepo": gitRepo,
			}).WithError(err).Error("Unable to update git repository")
			failed = append(failed, gitRepo.Name)
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
	for _, repo := range mgr.repos() {
		if repo.exists() {
			logrus.WithField("repo", repo).Debug("Repository already exists, skipping")
			continue
		}

		_, err := git.PlainClone(repo.path, false, &git.CloneOptions{
			URL: repo.Config.Remotes["origin"].URLs[0],
		})

		if err != nil {
			logrus.WithField("repo", repo.Name).WithError(err).Error("Failed to clone repository")
			failed = append(failed, repo.Name)
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("The following repos failed to update: %v", failed)
	}

	return nil
}

// Dump ...
func (mgr Manager) Dump() error {
	configFiles, err := dumpConfig()
	if err != nil {
		logrus.WithError(err).Error("Unable to find and save git configuration files")
	}

	symlinkMgr := symlink.NewManager(mgr.config)
	for _, f := range configFiles {
		symlinkMgr.Add(f, "")
	}

	repos, err := dumpRepos(mgr.reposDir())
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"reposDir": mgr.reposDir(),
		}).WithError(err).Error("Unable to list repos")
		return err
	}

	return file.SaveYaml(mgr.config.Fs, repos, mgr.config.Dotfiles, "repos")
}
