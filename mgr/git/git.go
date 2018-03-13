package git

import (
	"fmt"
	"path/filepath"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/file"
	"github.com/mbark/punkt/mgr/symlink"

	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
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

	for idx := range repos {
		repo := &repos[idx]
		worktree, err := mgr.config.Fs.Chroot(mgr.config.PunktHome + "/repos/" + repo.Name)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"repo": repo.Name,
			}).WithError(err).Error("Failed to chroot for repository")
			return []Repo{}
		}

		repo.worktree = worktree
	}

	return repos
}

// Update ...
func (mgr Manager) Update() error {
	failed := []string{}
	for _, repo := range mgr.repos() {
		err := repo.update(mgr.reposDir)
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

		if repo.exists() {
			logrus.WithField("repo", repo).Debug("Repository already exists, skipping")
			continue
		}

		// TODO: resolve how to use storer vs worktree (what do they mean?)
		dir, err := mgr.config.Fs.Chroot(mgr.reposDir + repo.Name)
		if err != nil {
			logger.WithError(err).Error("Failed to chroot to repo directory")
			failed = append(failed, repo.Name)
			continue
		}

		storer, err := filesystem.NewStorage(dir)
		if err != nil {
			logger.WithError(err).Error("Unable to create storage for repo")
			failed = append(failed, repo.Name)
			continue
		}

		remote := repo.Config.Remotes["origin"].URLs[0]
		logger.WithField("remote", remote).Debug("Cloning repository from remote")
		_, err = git.Clone(storer, repo.worktree, &git.CloneOptions{
			URL: remote,
		})

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
