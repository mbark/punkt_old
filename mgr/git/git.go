package git

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/file"
	"github.com/mbark/punkt/mgr/symlink"
	"github.com/mbark/punkt/run"

	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
)

var fileRegexp = regexp.MustCompile(`file\:(?P<File>.*?)\s.*`)

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

func (mgr Manager) reposDirectory() string {
	return filepath.Join(mgr.config.PunktHome, "repos")
}

func dirName(repo config.Config) string {
	s := strings.Split(repo.Remotes["origin"].URLs[0], "/")
	return s[len(s)-1]
}

func getRepo(cloneDir string) *git.Repository {
	logger := logrus.WithFields(logrus.Fields{
		"dir": cloneDir,
	})

	r, err := git.PlainOpen(cloneDir)
	if err != nil {
		logger.WithError(err).Error("Unable to open git repository")
		return nil
	}

	return r
}

// Update ...
func (mgr Manager) Update() {
	repos := mgr.repos()
	reposDir := mgr.reposDirectory()
	for _, repo := range repos {
		cloneDir := filepath.Join(reposDir, dirName(repo))

		logger := logrus.WithFields(logrus.Fields{
			"repo": repo,
			"dir":  cloneDir,
		})

		r, err := git.PlainOpen(cloneDir)
		if err != nil {
			logger.WithError(err).Fatal("Unable to open git repository")
		}

		w, err := r.Worktree()
		if err != nil {
			logger.WithError(err).Fatal("Unable to get working tree of git repository")
		}

		err = w.Pull(&git.PullOptions{RemoteName: "origin"})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			logger.WithError(err).Fatal("Unable to pull git repository")
		}
	}
}

func (mgr Manager) repos() []config.Config {
	repos := []config.Config{}
	file.Read(&repos, mgr.config.Dotfiles, "repos")

	return repos
}

// Ensure ...
func (mgr Manager) Ensure() {
	repos := []config.Config{}
	file.Read(&repos, mgr.config.Dotfiles, "repos")

	reposDir := mgr.reposDirectory()
	for _, repo := range repos {
		cloneDir := filepath.Join(reposDir, dirName(repo))
		if exists(repo, cloneDir) {
			logrus.WithFields(logrus.Fields{
				"dir":  cloneDir,
				"repo": repo,
			}).Debug("Repository already exists, skipping")
			continue
		}

		git.PlainClone(cloneDir, false, &git.CloneOptions{
			URL: repo.Remotes["origin"].URLs[0],
		})
	}
}

func exists(repo config.Config, path string) bool {
	r, err := git.PlainOpen(path)
	if err == git.ErrRepositoryNotExists {
		return false
	}

	if r != nil {
		return true
	}

	return false
}

// Dump ...
func (mgr Manager) Dump() {
	configFiles := mgr.dumpConfig()
	repos := mgr.dumpRepos()

	symlinkMgr := symlink.NewManager(mgr.config)

	for _, f := range configFiles {
		symlinkMgr.Add(f, "")
	}

	file.SaveYaml(repos, mgr.config.Dotfiles, "repos")
}

func (mgr Manager) dumpConfig() []string {
	// this is currently not suppported via the git library
	cmd := exec.Command("git", "config", "--list", "--show-origin", "--global")
	stdout := run.CaptureOutput(cmd)
	run.Run(cmd)

	output := strings.TrimSpace(stdout.String())
	rows := strings.Split(output, "\n")

	fileSet := make(map[string]struct{})

	for _, row := range rows {
		logrus.WithField("row", row).Info("Row")
		match := fileRegexp.FindStringSubmatch(row)
		logrus.WithField("match", match).Info("match")

		if len(match) > 1 {
			fileSet[match[1]] = struct{}{}
		}
	}

	files := []string{}
	for key := range fileSet {
		files = append(files, key)
	}

	return files
}

func (mgr Manager) dumpRepos() []config.Config {
	reposDir := mgr.reposDirectory()
	logger := logrus.WithField("reposDir", reposDir)
	files, err := ioutil.ReadDir(reposDir)

	if err != nil {
		logger.WithError(err).Fatal("Unable to list files in the repos directory")
	}

	repos := []config.Config{}
	for _, file := range files {
		if file.Mode()&os.ModeDir == 0 {
			continue
		}

		repo := getRepo(filepath.Join(reposDir, file.Name()))
		if repo == nil {
			continue
		}

		conf, err := repo.Config()
		if err != nil {
			logger.WithError(err).Error("Unable to get git repo config")
		}

		repos = append(repos, *conf)
	}

	return repos
}
