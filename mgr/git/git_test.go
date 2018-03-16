package git_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
	"time"

	g "github.com/onsi/ginkgo"
	m "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-billy.v4/osfs"
	"gopkg.in/src-d/go-billy.v4/util"
	goGit "gopkg.in/src-d/go-git.v4"
	goConfig "gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
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
			expected := []git.Repo{
				*addGitRepo(config, "repo1", nil),
				*addGitRepo(config, "repo2", nil),
			}

			unmarshalAndMarshal(&expected)
			m.Expect(mgr.Dump()).To(m.Succeed())

			actual := []git.Repo{}
			err := file.Read(config.Fs, &actual, config.Dotfiles, "repos")
			m.Expect(err).To(m.BeNil())
			m.Expect(actual).Should(m.ConsistOf(expected))
		})

		g.It("should fail if the repos directory contains non-git repos", func() {
			err := config.Fs.MkdirAll(config.PunktHome+"/repos/notGit", os.ModePerm)
			m.Expect(err).To(m.BeNil())
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
	})

	var _ = g.Context("when running Ensure", func() {
		g.It("should succeed when no repos file exists", func() {
			m.Expect(mgr.Ensure()).To(m.Succeed())
		})

		g.It("should do nothing if the repo already exists", func() {
			addGitRepo(config, "repo", nil)
			m.Expect(mgr.Dump()).To(m.Succeed())

			m.Expect(mgr.Ensure()).To(m.Succeed())
		})

		g.It("should fail if some repo can't be cloned", func() {
			dest := config.PunktHome + "/repos/"
			addGitRepo(config, "repo", nil)
			m.Expect(mgr.Dump()).To(m.Succeed())
			err := util.RemoveAll(config.Fs, dest+"repo")
			m.Expect(err).To(m.BeNil())

			m.Expect(mgr.Ensure()).NotTo(m.Succeed())
		})
	})

	var _ = g.Context("when cloning or fetching from another repo", func() {
		var tmpDir string

		g.BeforeEach(func() {
			var err error
			tmpDir, err = ioutil.TempDir("", "git-ensure")
			m.Expect(err).To(m.BeNil())
			config.Fs = osfs.New(tmpDir)

			mgr = git.NewManager(*config)

			addGitRepo(config, "cloneMe", &cloneConfig{
				doCommit: true,
				dest:     config.UserHome,
				remotes:  []string{"url"},
			})

			dest := config.PunktHome + "/repos/"
			addGitRepo(config, "repo", &cloneConfig{
				doCommit: false,
				remotes:  []string{tmpDir + config.UserHome + "/cloneMe"},
				dest:     dest,
			})
			m.Expect(mgr.Dump()).To(m.Succeed())
		})

		g.It("should clone the repositories specified", func() {
			err := util.RemoveAll(config.Fs, config.PunktHome+"/repos/repo")
			m.Expect(err).To(m.BeNil())
			m.Expect(mgr.Ensure()).To(m.Succeed())

			_, err = config.Fs.ReadDir(config.PunktHome + "/repos/repo")
			m.Expect(err).To(m.BeNil())
		})

		g.It("should do a fetch for the repositories", func() {
			m.Expect(mgr.Update()).To(m.Succeed())
		})
	})

	var _ = g.Context("when running Update", func() {
		g.It("should do nothing and succeed if no repos are cloned", func() {
			m.Expect(mgr.Update()).To(m.Succeed())
		})

		g.It("should fail if some repos can't be updated", func() {
			addGitRepo(config, "someRepo", nil)
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

type cloneConfig struct {
	doCommit bool
	dest     string
	remotes  []string
}

func unmarshalAndMarshal(out interface{}) {
	marshalled, err := yaml.Marshal(out)
	m.Expect(err).To(m.BeNil())
	err = yaml.Unmarshal(marshalled, out)
	m.Expect(err).To(m.BeNil())
}

func addGitRepo(config *conf.Config, name string, cloneConf *cloneConfig) *git.Repo {
	if cloneConf == nil {
		cloneConf = &cloneConfig{
			doCommit: true,
			dest:     config.PunktHome + "/repos",
			remotes:  []string{"/some/path/" + name},
		}
	}

	dir, err := config.Fs.Chroot(cloneConf.dest + "/" + name)
	m.Expect(err).To(m.BeNil())

	storage, err := filesystem.NewStorage(dir)
	m.Expect(err).To(m.BeNil())
	repo, err := goGit.Init(storage, dir)
	m.Expect(err).To(m.BeNil())

	_, err = repo.CreateRemote(&goConfig.RemoteConfig{
		Name:  "origin",
		Fetch: []goConfig.RefSpec{},
		URLs:  cloneConf.remotes,
	})
	m.Expect(err).To(m.BeNil())

	w, err := repo.Worktree()
	m.Expect(err).To(m.BeNil())

	if cloneConf.doCommit {
		f, err := dir.Create("afile")
		m.Expect(err).To(m.BeNil())
		_, err = w.Add(f.Name())
		m.Expect(err).To(m.BeNil())

		_, err = w.Commit("A commit", &goGit.CommitOptions{
			Author: &object.Signature{
				Name:  "John Doe",
				Email: "john@doe.org",
				When:  time.Now(),
			},
		})
		m.Expect(err).To(m.BeNil())
	}

	gitRepo, err := git.OpenRepo(dir, name)
	m.Expect(err).To(m.BeNil())
	return gitRepo
}
