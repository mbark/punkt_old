package git

import (
	"bytes"

	"github.com/BurntSushi/toml"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/mbark/punkt/mgr/symlink"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	gitconf "gopkg.in/src-d/go-git.v4/config"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/file"
)

// ErrRepositoryNotFoundInConfig ...
var ErrRepositoryNotFoundInConfig = errors.New("repository not found in config")

// Repo describes a git repository
type Repo struct {
	Name   string
	Path   string
	Config *gitconf.Config
}

// Manager ...
type Manager struct {
	LinkManager symlink.LinkManager
	RepoManager RepoManager
	config      conf.Config
	configFile  string
}

// Config ...
type Config struct {
	Symlinks     []symlink.Symlink
	Repositories []Repo
}

// NewManager ...
func NewManager(c conf.Config, configFile string) *Manager {
	return &Manager{
		LinkManager: symlink.NewLinkManager(c),
		RepoManager: NewRepoManager(c.Fs),
		config:      c,
		configFile:  configFile,
	}
}

func (mgr Manager) readConfig() Config {
	var config Config
	err := file.ReadToml(mgr.config.Fs, &config, mgr.configFile)
	if err == file.ErrNoSuchFile {
		return Config{}
	}

	return config
}

// Name ...
func (mgr Manager) Name() string {
	return "git"
}

// Add ...
func (mgr Manager) Add(path string) error {
	repo, err := mgr.RepoManager.Dump(path)
	if err != nil {
		return errors.Wrapf(err, "failed to dump repository at path: %s", path)
	}

	config := mgr.readConfig()
	config.Repositories = append(config.Repositories, *repo)
	return file.SaveToml(mgr.config.Fs, config, mgr.configFile)
}

// Remove ...
func (mgr Manager) Remove(path string) error {
	config := mgr.readConfig()

	index := -1
	for i, repo := range config.Repositories {
		if repo.Path == path {
			index = i
		}
	}

	if index < 0 {
		logrus.WithFields(logrus.Fields{
			"path":   path,
			"config": config,
		}).Error("repository not found in config file")
		return ErrRepositoryNotFoundInConfig
	}

	config.Repositories = append(config.Repositories[:index], config.Repositories[index+1:]...)
	return file.SaveToml(mgr.config.Fs, config, mgr.configFile)
}

// Update ...
func (mgr Manager) Update() error {
	var result error
	for _, repo := range mgr.readConfig().Repositories {
		_, err := mgr.RepoManager.Update(repo.Path)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"repo": repo,
			}).WithError(err).Error("Unable to update git repository")
			result = multierror.Append(result, err)
		}
	}

	return result
}

// Ensure ...
func (mgr Manager) Ensure() error {
	var result error
	for _, repo := range mgr.readConfig().Repositories {
		err := mgr.RepoManager.Ensure(repo)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"repo": repo,
			}).WithError(err).Error("Failed to ensure git repository")
			result = multierror.Append(result, err)
		}
	}

	return result
}

// Dump ...
func (mgr Manager) Dump() (string, error) {
	configFiles := globalConfigFiles(mgr.config.Command)

	var symlinks []symlink.Symlink
	for _, f := range configFiles {
		s := mgr.LinkManager.New("", f)
		unexpanded := mgr.LinkManager.Unexpand(*s)
		symlinks = append(symlinks, *unexpanded)

		logrus.WithFields(logrus.Fields{
			"configFile": f,
			"symlink":    s,
		}).Debug("Storing symlink to config file")
	}

	config := Config{
		Symlinks:     symlinks,
		Repositories: []Repo{},
	}

	var out bytes.Buffer
	encoder := toml.NewEncoder(&out)
	err := encoder.Encode(config)

	return out.String(), errors.Wrap(err, "failed to encode git-configuration")
}
