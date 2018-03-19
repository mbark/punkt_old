package git_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	g "github.com/onsi/ginkgo"
	m "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-billy.v4/util"
	yaml "gopkg.in/yaml.v2"

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
	ensurer func(string, git.Repo) error
	updater func(string) (bool, error)
}

func (mgr fakeRepoManager) Dump(dir string) (*git.Repo, error) {
	if mgr.dumper != nil {
		return mgr.dumper(dir)
	}

	return &git.Repo{Name: filepath.Base(dir), Config: nil}, nil
}

func (mgr fakeRepoManager) Ensure(dir string, repo git.Repo) error {
	if mgr.ensurer != nil {
		return mgr.ensurer(dir, repo)
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

	g.BeforeEach(func() {
		config = &conf.Config{
			UserHome:   "/home",
			PunktHome:  "/home/.config/punkt",
			Dotfiles:   "/home/.dotfiles",
			Fs:         memfs.New(),
			WorkingDir: "/home",
			Command:    fakeCommand,
		}

		mgr = git.NewManager(*config)
		repoMgr = &fakeRepoManager{}
		mgr.RepoManager = repoMgr
		logrus.SetLevel(logrus.PanicLevel)
	})

	var _ = g.Context("when running Dump", func() {
		g.It("should fail if the command fails", func() {
			config.Command = fakeWithEnvCommand("FAILING=true")
			mgr = git.NewManager(*config)
			m.Expect(mgr.Dump()).NotTo(m.Succeed())
		})

		g.It("should create symlinks for the config files in the git config", func() {
			config.Command = fakeWithEnvCommand("WITH_GITCONFIG=true")
			mgr = git.NewManager(*config)

			_, err := config.Fs.Create(config.UserHome + "/.gitconfig")
			m.Expect(err).To(m.BeNil())
			_, err = config.Fs.Create(config.UserHome + "/.config/git/config")
			m.Expect(err).To(m.BeNil())

			expected := []symlink.Symlink{
				*symlink.NewSymlink(config.Fs, "~/.dotfiles/.gitconfig", "~/.gitconfig"),
				*symlink.NewSymlink(config.Fs, "~/.dotfiles/.config/git/config", "~/.config/git/config"),
			}
			unmarshalAndMarshal(&expected)

			m.Expect(mgr.Dump()).To(m.Succeed())

			actual := []symlink.Symlink{}
			err = file.Read(config.Fs, &actual, config.Dotfiles, "symlinks")
			m.Expect(err).To(m.BeNil())

			m.Expect(actual).Should(m.ConsistOf(expected))
		})

		g.It("should fail if some symlink can't be created", func() {
			config.Command = fakeWithEnvCommand("WITH_GITCONFIG=true")
			mgr = git.NewManager(*config)

			m.Expect(mgr.Dump()).NotTo(m.Succeed())

			actual := []symlink.Symlink{}
			err := file.Read(config.Fs, &actual, config.Dotfiles, "symlinks")
			m.Expect(err).NotTo(m.BeNil())
			m.Expect(actual).To(m.Equal([]symlink.Symlink{}))
		})

		g.It("should dump the repos cloned in the repos directory", func() {
			repos := []string{"repo1", "repo2"}
			expected := []git.Repo{}
			for _, r := range repos {
				addFakeRepo(config, r)
				expected = append(expected, git.Repo{Name: r})
			}

			m.Expect(mgr.Dump()).To(m.Succeed())

			actual := []git.Repo{}
			err := file.Read(config.Fs, &actual, config.Dotfiles, "repos")
			m.Expect(err).To(m.BeNil())
			m.Expect(actual).Should(m.ConsistOf(expected))
		})

		g.It("should fail if some repo can't be dumped", func() {
			repos := []string{"repo1", "repo2"}
			for _, r := range repos {
				addFakeRepo(config, r)
			}

			repoMgr.dumper = func(dir string) (*git.Repo, error) {
				if strings.HasSuffix(dir, repos[1]) {
					return nil, fmt.Errorf("can't dummp %s", dir)
				}

				return &git.Repo{Name: dir}, nil
			}

			m.Expect(mgr.Dump()).NotTo(m.Succeed())
		})

		g.It("should fail if the repos directory contains non-git repos", func() {
			repoMgr.dumper = func(dir string) (*git.Repo, error) {
				return nil, fmt.Errorf("error")
			}
			addFakeRepo(config, "notGit")

			m.Expect(mgr.Dump()).NotTo(m.Succeed())
		})

		g.It("should ignore non-directories in the repos directory", func() {
			_, err := config.Fs.Create(config.PunktHome + "/repos/file")
			m.Expect(err).To(m.BeNil())
			m.Expect(mgr.Dump()).To(m.Succeed())

			actual := []git.Repo{}
			err = file.Read(config.Fs, &actual, config.Dotfiles, "repos")
			m.Expect(err).To(m.BeNil())
			m.Expect(actual).Should(m.ConsistOf([]git.Repo{}))
		})

		g.It("should append the repos directory to the path", func() {
			addFakeRepo(config, "repo")

			m.Expect(mgr.Dump()).To(m.Succeed())
			repoMgr.dumper = func(dir string) (*git.Repo, error) {
				expected := filepath.Join(config.PunktHome, "repos", "repo")
				m.Expect(dir).To(m.Equal(expected))
				return &git.Repo{Name: dir}, nil
			}

			m.Expect(mgr.Dump()).To(m.Succeed())
		})
	})

	var _ = g.Context("when running Ensure", func() {
		g.It("should succeed when no repos file exists", func() {
			m.Expect(mgr.Ensure()).To(m.Succeed())
		})

		g.It("should do nothing if the repo already exists", func() {
			m.Expect(mgr.Dump()).To(m.Succeed())

			m.Expect(mgr.Ensure()).To(m.Succeed())
		})

		g.It("should fail if some repo can't be cloned", func() {
			repoMgr.ensurer = func(dir string, repo git.Repo) error {
				return fmt.Errorf("fail")
			}
			dir := addFakeRepo(config, "repo")
			m.Expect(mgr.Dump()).To(m.Succeed())
			err := util.RemoveAll(config.Fs, dir)
			m.Expect(err).To(m.BeNil())

			m.Expect(mgr.Ensure()).NotTo(m.Succeed())
		})
	})

	var _ = g.Context("when running Update", func() {
		g.It("should do nothing and succeed if no repos are cloned", func() {
			m.Expect(mgr.Update()).To(m.Succeed())
		})

		g.It("should succeed if the repo can be updated", func() {
			addFakeRepo(config, "repo")
			m.Expect(mgr.Dump()).To(m.Succeed())

			m.Expect(mgr.Update()).To(m.Succeed())
		})

		g.It("should fail if some repos can't be updated", func() {
			repoMgr.updater = func(dir string) (bool, error) {
				return false, fmt.Errorf("fail")
			}
			addFakeRepo(config, "repo")
			m.Expect(mgr.Dump()).To(m.Succeed())

			m.Expect(mgr.Update()).NotTo(m.Succeed())
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

func unmarshalAndMarshal(out interface{}) {
	marshalled, err := yaml.Marshal(out)
	m.Expect(err).To(m.BeNil())
	err = yaml.Unmarshal(marshalled, out)
	m.Expect(err).To(m.BeNil())
}

func addFakeRepo(config *conf.Config, name string) string {
	dir := filepath.Join(config.PunktHome, "repos", name)
	err := config.Fs.MkdirAll(dir, os.ModePerm)
	m.Expect(err).To(m.BeNil())

	return dir
}
