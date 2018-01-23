package git

import (
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
	file.Read(&repos, mgr.config.Dotfiles, "repos")

	return repos
}

// Update ...
func (mgr Manager) Update() {
	for _, gitRepo := range mgr.repos() {
		err := gitRepo.update(mgr.reposDir())
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"gitRepo": gitRepo,
			}).WithError(err).Error("Unable to update git repository")
		}
	}
}

// Ensure ...
func (mgr Manager) Ensure() {
	for _, repo := range mgr.repos() {
		if repo.exists() {
			logrus.WithField("repo", repo).Debug("Repository already exists, skipping")
			continue
		}

		git.PlainClone(repo.path, false, &git.CloneOptions{
			URL: repo.Config.Remotes["origin"].URLs[0],
		})
	}
}

// Dump ...
func (mgr Manager) Dump() {
	configFiles := dumpConfig()

	symlinkMgr := symlink.NewManager(mgr.config)
	for _, f := range configFiles {
		symlinkMgr.Add(f, "")
	}

	repos, err := dumpRepos(mgr.reposDir())
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"reposDir": mgr.reposDir(),
		}).WithError(err).Error("Unable to list repos")
	} else {
		file.SaveYaml(repos, mgr.config.Dotfiles, "repos")
	}
}
