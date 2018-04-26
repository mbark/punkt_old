package git_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
	g "github.com/onsi/ginkgo"
	m "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-billy.v4/memfs"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/file"
	"github.com/mbark/punkt/mgr/git"
	"github.com/mbark/punkt/mgr/symlink"
)

func TestGit(t *testing.T) {
	m.RegisterFailHandler(g.Fail)
	g.RunSpecs(t, "Git Suite")
}

type fakeRepoManager struct {
	dumper  func(string) (*git.Repo, error)
	ensurer func(git.Repo) error
	updater func(string) (bool, error)
}

func (mgr fakeRepoManager) Dump(dir string) (*git.Repo, error) {
	if mgr.dumper != nil {
		return mgr.dumper(dir)
	}

	return &git.Repo{Name: filepath.Base(dir), Config: nil}, nil
}

func (mgr fakeRepoManager) Ensure(repo git.Repo) error {
	if mgr.ensurer != nil {
		return mgr.ensurer(repo)
	}

	return nil
}

func (mgr fakeRepoManager) Update(dir string) (bool, error) {
	if mgr.updater != nil {
		return mgr.updater(dir)
	}

	return true, nil
}

var _ = g.Describe("Git: Manager", func() {
	var config *conf.Config
	var mgr *git.Manager
	var repoMgr *fakeRepoManager
	var configFile string

	g.BeforeEach(func() {
		config = &conf.Config{
			UserHome:   "/home",
			PunktHome:  "/home/.config/punkt",
			Dotfiles:   "/home/.dotfiles",
			Fs:         memfs.New(),
			WorkingDir: "/home",
			Command:    fakeCommand,
		}

		configFile = filepath.Join(config.PunktHome, "git.toml")
		mgr = git.NewManager(*config, configFile)
		repoMgr = &fakeRepoManager{}
		mgr.RepoManager = repoMgr
		logrus.SetLevel(logrus.PanicLevel)
	})

	g.It("should be called git", func() {
		m.Expect(mgr.Name()).To(m.Equal("git"))
	})

	var _ = g.Context("when running Dump", func() {
		g.It("should return valid toml", func() {
			dumped, err := mgr.Dump()
			m.Expect(err).To(m.BeNil())

			var actual git.Config
			_, err = toml.Decode(dumped, &actual)
			m.Expect(err).To(m.BeNil())

			m.Expect(actual.Symlinks).Should(m.BeEmpty())
			m.Expect(actual.Repositories).Should(m.BeEmpty())
		})

		g.It("should contain the files to symlink", func() {
			config.Command = fakeWithEnvCommand("WITH_GITCONFIG=true")
			mgr = git.NewManager(*config, configFile)

			expected := []symlink.Symlink{
				{Target: "~/.dotfiles/.gitconfig", Link: "~/.gitconfig"},
				{Target: "~/.dotfiles/.config/git/config", Link: "~/.config/git/config"},
			}

			dumped, err := mgr.Dump()
			m.Expect(err).To(m.BeNil())

			var actual git.Config
			_, err = toml.Decode(dumped, &actual)
			m.Expect(err).To(m.BeNil())

			m.Expect(actual.Symlinks).Should(m.ConsistOf(expected))
		})
	})

	var _ = g.Context("when running Ensure", func() {
		g.It("should succeed when no repos file exists", func() {
			m.Expect(mgr.Ensure()).To(m.Succeed())
		})

		g.It("should do nothing if the repo already exists", func() {
			dir := addFakeRepo(config, "repo")
			m.Expect(mgr.Add(dir)).To(m.Succeed())

			m.Expect(mgr.Ensure()).To(m.Succeed())
		})

		g.It("should fail if some repos can't be ensured", func() {
			repoMgr.ensurer = func(repo git.Repo) error {
				return fmt.Errorf("fail")
			}
			dir := addFakeRepo(config, "repo")
			m.Expect(mgr.Add(dir)).To(m.Succeed())

			m.Expect(mgr.Ensure()).NotTo(m.Succeed())
		})
	})

	var _ = g.Context("when running Update", func() {
		g.It("should do nothing and succeed if no repos are cloned", func() {
			m.Expect(mgr.Update()).To(m.Succeed())
		})

		g.It("should succeed if the repo can be updated", func() {
			addFakeRepo(config, "repo")
			_, err := mgr.Dump()
			m.Expect(err).To(m.BeNil())

			m.Expect(mgr.Update()).To(m.Succeed())
		})

		g.It("should fail if some repos can't be updated", func() {
			repoMgr.updater = func(dir string) (bool, error) {
				return false, fmt.Errorf("fail")
			}
			dir := addFakeRepo(config, "repo")
			m.Expect(mgr.Add(dir)).To(m.Succeed())

			m.Expect(mgr.Update()).NotTo(m.Succeed())
		})
	})

	var _ = g.Context("when getting Symlinks", func() {
		g.It("should return the saved symlinks", func() {
			expected := []symlink.Symlink{
				{Target: "~/.dotfiles/.gitconfig", Link: "~/.gitconfig"},
				{Target: "~/.dotfiles/.config/git/config", Link: "~/.config/git/config"},
			}

			err := file.SaveToml(config.Fs, &git.Config{Symlinks: expected}, configFile)
			m.Expect(err).To(m.BeNil())

			actual, err := mgr.Symlinks()
			m.Expect(err).To(m.BeNil())

			m.Expect(actual).Should(m.ConsistOf(expected))
		})

		g.It("should return an empty list if the config doesn't exit", func() {
			actual, err := mgr.Symlinks()
			m.Expect(err).To(m.BeNil())
			m.Expect(actual).To(m.BeEmpty())
		})

		g.It("should return an error if the file can't be read", func() {
			err := file.Save(config.Fs, "foo", configFile)
			m.Expect(err).To(m.BeNil())

			actual, err := mgr.Symlinks()
			m.Expect(actual).To(m.BeNil())
			m.Expect(err).NotTo(m.BeNil())
		})
	})

	var _ = g.Context("when removing a git repo", func() {
		g.It("should be possible to remove a repo", func() {
			repoPath := filepath.Join(config.UserHome, "repo")

			c := git.Config{Repositories: []git.Repo{{Path: repoPath}}}
			err := file.SaveToml(config.Fs, &c, configFile)
			m.Expect(err).To(m.BeNil())

			m.Expect(mgr.Remove(repoPath)).To(m.Succeed())

			var actual git.Config
			err = file.ReadToml(config.Fs, &actual, configFile)
			m.Expect(err).To(m.BeNil())

			m.Expect(actual.Repositories).To(m.BeEmpty())
		})

		g.It("should return an error if the repo doesn't exist", func() {
			repoPath := filepath.Join(config.UserHome, "repo")
			c := git.Config{Repositories: []git.Repo{{Path: repoPath}}}
			err := file.SaveToml(config.Fs, &c, configFile)
			m.Expect(err).To(m.BeNil())

			err = mgr.Remove("/non/existant")
			m.Expect(err).To(m.Equal(git.ErrRepositoryNotFoundInConfig))
		})
	})
})

func fakeWithEnvCommand(env string) func(string, ...string) *exec.Cmd {
	return func(command string, args ...string) *exec.Cmd {
		cmd := fakeCommand(command, args...)
		cmd.Env = append(cmd.Env, env)
		return cmd
	}
}

func fakeCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestGitHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestGitHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	if os.Getenv("FAILING") == "true" {
		os.Exit(3)
	}

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	cmd, args := args[0], args[1:]

	switch cmd {
	case "git":
		if args[0] != "config" {
			os.Exit(1)
		}

		if os.Getenv("WITH_GITCONFIG") == "true" {
			fmt.Println(gitConfig)
		} else {
			fmt.Println(``)
		}

		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "non-brew command called\n")
		os.Exit(1)
	}

	os.Exit(1)
}

const gitConfig = `
file:/home/.gitconfig           user.email=user.name@mail.io
file:/home/.gitconfig           user.name=User Name
file:/home/.config/git/config   push.default=simple
`

func addFakeRepo(config *conf.Config, name string) string {
	dir := filepath.Join(config.PunktHome, "repos", name)
	err := config.Fs.MkdirAll(dir, os.ModePerm)
	m.Expect(err).To(m.BeNil())

	return dir
}
