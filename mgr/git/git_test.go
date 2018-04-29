package git_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"gopkg.in/src-d/go-billy.v4/memfs"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/file"
	"github.com/mbark/punkt/mgr/git"
	"github.com/mbark/punkt/mgr/symlink"
)

func TestGit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Git Suite")
}

type mockRepoManager struct {
	mock.Mock
}

func (m *mockRepoManager) Dump(dir string) (*git.Repo, error) {
	args := m.Called(dir)
	return args.Get(0).(*git.Repo), args.Error(1)
}

func (m *mockRepoManager) Ensure(repo git.Repo) error {
	args := m.Called(repo)
	return args.Error(0)
}

func (m *mockRepoManager) Update(dir string) (bool, error) {
	args := m.Called(dir)
	return args.Bool(0), args.Error(1)
}

var _ = Describe("Git: Manager", func() {
	var config *conf.Config
	var mgr *git.Manager
	var repoMgr *mockRepoManager
	var configFile string

	BeforeEach(func() {
		logrus.SetLevel(logrus.PanicLevel)

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

		repoMgr = new(mockRepoManager)
		mgr.RepoManager = repoMgr

		repoMgr.On("Dump", mock.Anything).Return(new(git.Repo), nil)
	})

	It("should be called git", func() {
		Expect(mgr.Name()).To(Equal("git"))
	})

	var _ = Context("Dump", func() {
		It("should return valid toml", func() {
			dumped, err := mgr.Dump()
			Expect(err).To(BeNil())

			var actual git.Config
			_, err = toml.Decode(dumped, &actual)
			Expect(err).To(BeNil())

			Expect(actual.Symlinks).Should(BeEmpty())
			Expect(actual.Repositories).Should(BeEmpty())
		})

		It("should contain the files to symlink", func() {
			config.Command = fakeWithEnvCommand("WITH_GITCONFIG=true")
			mgr = git.NewManager(*config, configFile)

			expected := []symlink.Symlink{
				{Target: "~/.dotfiles/.gitconfig", Link: "~/.gitconfig"},
				{Target: "~/.dotfiles/.config/git/config", Link: "~/.config/git/config"},
			}

			dumped, err := mgr.Dump()
			Expect(err).To(BeNil())

			var actual git.Config
			_, err = toml.Decode(dumped, &actual)
			Expect(err).To(BeNil())

			Expect(actual.Symlinks).Should(ConsistOf(expected))
		})

		It("should return no symlinks if finding the config files fails", func() {
			config.Command = fakeWithEnvCommand("FAILING=true")
			mgr = git.NewManager(*config, configFile)

			dumped, err := mgr.Dump()
			Expect(err).To(BeNil())

			var actual git.Config
			_, err = toml.Decode(dumped, &actual)
			Expect(err).To(BeNil())

			Expect(actual.Symlinks).Should(BeEmpty())
		})
	})

	var _ = Context("Ensure", func() {
		It("should succeed when no repos file exists", func() {
			Expect(mgr.Ensure()).To(Succeed())
		})

		It("should do nothing if the repo already exists", func() {
			repoMgr.On("Ensure", mock.Anything).Return(nil)
			dir := addFakeRepo(config, "repo")
			Expect(mgr.Add(dir)).To(Succeed())

			Expect(mgr.Ensure()).To(Succeed())
		})

		It("should fail if some repos can't be ensured", func() {
			repoMgr.On("Ensure", mock.Anything).Return(fmt.Errorf("fail"))
			dir := addFakeRepo(config, "repo")
			Expect(mgr.Add(dir)).To(Succeed())

			Expect(mgr.Ensure()).NotTo(Succeed())
		})
	})

	var _ = Context("Update", func() {
		It("should do nothing and succeed if no repos are cloned", func() {
			Expect(mgr.Update()).To(Succeed())
		})

		It("should succeed if the repo can be updated", func() {
			addFakeRepo(config, "repo")
			_, err := mgr.Dump()
			Expect(err).To(BeNil())

			Expect(mgr.Update()).To(Succeed())
		})

		It("should fail if some repos can't be updated", func() {
			repoMgr.On("Update", mock.Anything).Return(false, fmt.Errorf("fail"))
			dir := addFakeRepo(config, "repo")
			Expect(mgr.Add(dir)).To(Succeed())

			Expect(mgr.Update()).NotTo(Succeed())
		})
	})

	var _ = Context("when removing a git repo", func() {
		It("should be possible to remove a repo", func() {
			repoPath := filepath.Join(config.UserHome, "repo")

			c := git.Config{Repositories: []git.Repo{{Path: repoPath}}}
			err := file.SaveToml(config.Fs, &c, configFile)
			Expect(err).To(BeNil())

			Expect(mgr.Remove(repoPath)).To(Succeed())

			var actual git.Config
			err = file.ReadToml(config.Fs, &actual, configFile)
			Expect(err).To(BeNil())

			Expect(actual.Repositories).To(BeEmpty())
		})

		It("should return an error if the repo doesn't exist", func() {
			repoPath := filepath.Join(config.UserHome, "repo")
			c := git.Config{Repositories: []git.Repo{{Path: repoPath}}}
			err := file.SaveToml(config.Fs, &c, configFile)
			Expect(err).To(BeNil())

			err = mgr.Remove("/non/existant")
			Expect(err).To(Equal(git.ErrRepositoryNotFoundInConfig))
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
	Expect(err).To(BeNil())

	return dir
}
