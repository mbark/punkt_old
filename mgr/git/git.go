package git

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/file"
	"github.com/mbark/punkt/mgr/symlink"
	"github.com/mbark/punkt/run"
	gitconf "gopkg.in/src-d/go-git.v4/config"

	"github.com/sirupsen/logrus"
)

var gitConfigFile = regexp.MustCompile(`file\:(?P<File>.*?)\s.*`)

// Repo describes a git repository
type Repo struct {
	Name   string
	Config *gitconf.Config
}

// Manager ...
type Manager struct {
	RepoManager RepoManager
	config      conf.Config
	reposDir    string
}

// NewManager ...
func NewManager(c conf.Config) *Manager {
	return &Manager{
		RepoManager: NewGoGitRepoManager(c.Fs),
		config:      c,
		reposDir:    filepath.Join(c.PunktHome, "repos"),
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
		dir := filepath.Join(mgr.reposDir, repo.Name)
		_, err := mgr.RepoManager.Update(dir)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"repo": repo,
				"dir":  dir,
			}).WithError(err).Error("Unable to update git repository")
			failed = append(failed, repo.Name)
			continue
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
		dir := filepath.Join(mgr.reposDir, repo.Name)
		err := mgr.RepoManager.Ensure(dir, repo)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"repo": repo,
				"dir":  dir,
			}).WithError(err).Error("Failed to ensure git repository")
			failed = append(failed, repo.Name)
			continue
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("The following repos failed to update: %v", failed)
	}

	return nil
}

// Dump ...
func (mgr Manager) Dump() error {
	configFiles, err := mgr.globalConfigFiles()
	if err != nil {
		logrus.WithError(err).Error("Unable to find and save git configuration files")
		return err
	}

	symlinkMgr := symlink.NewManager(mgr.config)
	for _, f := range configFiles {
		_, err := symlinkMgr.Add(f)
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

func (mgr Manager) globalConfigFiles() ([]string, error) {
	// this is currently not suppported via the git library
	cmd := mgr.config.Command("git", "config", "--list", "--show-origin", "--global")
	stdout, stderr := run.CaptureOutput(cmd)
	err := run.Run(cmd)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"stdout": stdout.String(),
			"stderr": stderr.String(),
		}).WithError(err).Error("Failed to run git config")
		return []string{}, err
	}

	logrus.WithFields(logrus.Fields{
		"stdout": stdout.String(),
	}).Debug("Got git config list successfully")

	output := strings.TrimSpace(stdout.String())
	rows := strings.Split(output, "\n")

	fileSet := make(map[string]struct{})

	for _, row := range rows {
		match := gitConfigFile.FindStringSubmatch(row)
		if len(match) > 1 {
			fileSet[match[1]] = struct{}{}
		}
	}

	files := []string{}
	for key := range fileSet {
		files = append(files, key)
	}

	return files, nil
}

func (mgr Manager) dumpRepos() ([]Repo, error) {
	repos := []Repo{}

	files, err := mgr.config.Fs.ReadDir(mgr.reposDir)
	if err != nil {
		logrus.WithError(err).Warning("Unable to read repos directory")
		return repos, err
	}

	for _, file := range files {
		if file.Mode()&os.ModeDir == 0 {
			continue
		}

		repo, err := mgr.RepoManager.Dump(filepath.Join(mgr.reposDir, file.Name()))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"repo": file.Name(),
			}).WithError(err).Warning("Unable to open git repository")
			return repos, err
		}

		repos = append(repos, *repo)
	}

	logrus.WithFields(logrus.Fields{
		"repos": repos,
	}).Debug("Found git repos to save")
	return repos, nil
}
