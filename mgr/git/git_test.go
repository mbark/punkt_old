package git_test

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	g "github.com/onsi/ginkgo"
	m "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-billy.v4/memfs"
	goGit "gopkg.in/src-d/go-git.v4"
	goConfig "gopkg.in/src-d/go-git.v4/config"
	goStorage "gopkg.in/src-d/go-git.v4/storage/filesystem"
	yaml "gopkg.in/yaml.v2"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/file"
	"github.com/mbark/punkt/mgr"
	"github.com/mbark/punkt/mgr/git"
	"github.com/mbark/punkt/mgr/symlink"
)

func TestGit(t *testing.T) {
	m.RegisterFailHandler(g.Fail)
	g.RunSpecs(t, "Git Suite")
}

var _ = g.Describe("Git: Manager", func() {
	var config *conf.Config
	var mgr mgr.Manager

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
				*symlink.NewSymlink(nil, "~/.dotfiles/.gitconfig", "~/.gitconfig"),
				*symlink.NewSymlink(nil, "~/.dotfiles/.config/git/config", "~/.config/git/config"),
			}
			m.Expect(err).To(m.BeNil())

			m.Expect(mgr.Dump()).To(m.Succeed())

			actual := []symlink.Symlink{}
			err = file.Read(config.Fs, &actual, config.Dotfiles, "symlinks")
			m.Expect(err).To(m.BeNil())

			// Due to the fact that symlinks are stored in an array and array
			// order isn't relevant we have to use this comparison check instead
			// of MatchYAML
			m.Expect(actual).Should(m.ConsistOf(expected[0], expected[1]))
		})

		g.It("should dump the repos cloned in the repos directory", func() {
			expected, err := yaml.Marshal([]git.Repo{
				*addGitRepo(config, "repo1"),
				*addGitRepo(config, "repo2"),
			})
			m.Expect(err).To(m.BeNil())
			m.Expect(mgr.Dump()).To(m.Succeed())

			actual, err := file.ReadAsString(config.Fs, config.Dotfiles, "repos")
			m.Expect(err).To(m.BeNil())
			m.Expect(actual).Should(m.MatchYAML(expected))
		})
	})

	g.It("should succeed when running Ensure without yaml file", func() {
		m.Expect(mgr.Ensure()).To(m.Succeed())
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

		return
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

func addGitRepo(config *conf.Config, name string) *git.Repo {
	dir, err := config.Fs.Chroot(config.PunktHome + "/repos/" + name)
	m.Expect(err).To(m.BeNil())
	storage, err := goStorage.NewStorage(dir)
	m.Expect(err).To(m.BeNil())
	repo, err := goGit.Init(storage, dir)
	m.Expect(err).To(m.BeNil())
	_, err = repo.CreateRemote(&goConfig.RemoteConfig{
		Name:  "origin",
		Fetch: []goConfig.RefSpec{},
		URLs:  []string{"url"},
	})
	m.Expect(err).To(m.BeNil())

	fs, err := config.Fs.Chroot(config.PunktHome + "/repos/" + name)
	m.Expect(err).To(m.BeNil())
	gitRepo, err := git.NewRepo(fs, "")
	m.Expect(err).To(m.BeNil())
	return gitRepo
}
