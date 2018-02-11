package brew_test

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	g "github.com/onsi/ginkgo"
	m "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-billy.v4/memfs"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/mgr"
	"github.com/mbark/punkt/mgr/brew"
)

func TestBrew(t *testing.T) {
	m.RegisterFailHandler(g.Fail)
	g.RunSpecs(t, "Brew Suite")
}

var _ = g.Describe("Brew", func() {
	var config *conf.Config
	var mgr mgr.Manager
	var brewfile string

	g.BeforeEach(func() {
		logrus.SetLevel(logrus.PanicLevel)
		config = &conf.Config{
			UserHome:   "/home",
			PunktHome:  "/home/.config/punkt",
			Dotfiles:   "/home/.dotfiles",
			Fs:         memfs.New(),
			WorkingDir: "/home",
			Command:    fakeCommand,
		}

		brewfile = config.UserHome + "/.Brewfile"
		_, err := config.Fs.Create(brewfile)
		m.Expect(err).To(m.BeNil())

		mgr = brew.NewManager(*config)
	})

	g.It("should call with expected format for 'ensure'", func() {
		m.Expect(mgr.Ensure()).To(m.Succeed())
	})

	g.It("should call with expected format for 'update'", func() {
		m.Expect(mgr.Update()).To(m.Succeed())
	})

	g.It("should succeed if the Brewfile is created", func() {
		m.Expect(mgr.Dump()).To(m.Succeed())
	})

	g.It("should fail to dump if the Brewfile doesn't exist", func() {
		m.Expect(config.Fs.Remove(brewfile)).To(m.Succeed())
		m.Expect(mgr.Dump()).NotTo(m.Succeed())
	})

	g.It("should symlink the Brewfile to the dotfiles directory", func() {
		m.Expect(mgr.Dump()).To(m.Succeed())
		path, err := config.Fs.Readlink(brewfile)
		m.Expect(err).To(m.BeNil())
		m.Expect(path).To(m.HavePrefix(config.Dotfiles))
	})

	g.Context("when 'brew bundle' fails", func() {
		g.BeforeEach(func() {
			config.Command = fakeFailingCommand
			mgr = brew.NewManager(*config)

			brewfile = config.UserHome + "/.Brewfile"
			_, err := config.Fs.Create(brewfile)
			m.Expect(err).To(m.BeNil())
		})

		g.It("should return an error for all commands", func() {
			m.Expect(mgr.Dump()).NotTo(m.Succeed())
			m.Expect(mgr.Update()).NotTo(m.Succeed())
			m.Expect(mgr.Ensure()).NotTo(m.Succeed())
		})

		g.It("should not attempt to symlink the Brewfile", func() {
			m.Expect(mgr.Dump()).NotTo(m.Succeed())
			_, err := config.Fs.Readlink(brewfile)
			m.Expect(err).NotTo(m.BeNil())
		})
	})
})

func fakeFailingCommand(command string, args ...string) *exec.Cmd {
	cmd := fakeCommand(command, args...)
	cmd.Env = append(cmd.Env, "FAILING=true")
	return cmd
}

func fakeCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestBrewHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestBrewHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

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
	case "brew":
		if args[0] != "bundle" || args[len(args)-1] != "--global" {
			fmt.Fprintf(os.Stderr, "Didn't call 'brew bundle ... --global': %v", args)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "non-brew command called\n")
		os.Exit(1)
	}

	switch args[1] {
	case "dump":
	case "--no-upgrade":
	case "--global":
	default:
		fmt.Fprintf(os.Stderr, "unexpected flag/command provided: %v", args)
		os.Exit(1)
	}

	if os.Getenv("FAILING") == "true" {
		os.Exit(3)
		return
	}
}
