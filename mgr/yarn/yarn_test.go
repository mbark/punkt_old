package yarn_test

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	g "github.com/onsi/ginkgo"
	m "github.com/onsi/gomega"
	"gopkg.in/src-d/go-billy.v4/memfs"

	"io/ioutil"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/mgr"
	"github.com/mbark/punkt/mgr/yarn"
)

func TestYarn(t *testing.T) {
	m.RegisterFailHandler(g.Fail)
	g.RunSpecs(t, "Yarn Suite")
}

var _ = g.Describe("Yarn", func() {
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

		mgr = yarn.NewManager(*config)
	})

	g.It("should call with expected format for 'ensure'", func() {
		m.Expect(mgr.Ensure()).To(m.Succeed())
	})

	g.It("should call with expected format for 'update'", func() {
		m.Expect(mgr.Update()).To(m.Succeed())
	})

	g.It("should fail when the symlinks for dump aren't created", func() {
		m.Expect(mgr.Dump()).NotTo(m.Succeed())
	})

	g.It("should fail to run 'ensure' if yarn global dir fails", func() {
		config.Command = fakeFailCommand("dir")
		mgr = yarn.NewManager(*config)
		m.Expect(mgr.Ensure()).NotTo(m.Succeed())
	})

	g.It("should fail to run 'ensure' if 'yarn' fails", func() {
		config.Command = fakeFailCommand("yarn")
		mgr = yarn.NewManager(*config)
		m.Expect(mgr.Ensure()).NotTo(m.Succeed())
	})

	g.It("should fail to run 'update' if 'yarn global upgrade' fails", func() {
		config.Command = fakeFailCommand("upgrade")
		mgr = yarn.NewManager(*config)
		m.Expect(mgr.Update()).NotTo(m.Succeed())
	})

	g.It("should succeed the dump if all symlinks are created", func() {
		mgr := yarn.NewManager(*config)

		for _, f := range mgr.ConfigFiles() {
			_, err := config.Fs.Create(f)
			m.Expect(err).To(m.BeNil())
		}
		m.Expect(mgr.Dump()).To(m.Succeed())
	})
})

func fakeFailCommand(fail string) func(string, ...string) *exec.Cmd {
	return func(command string, args ...string) *exec.Cmd {
		cmd := fakeCommand(command, args...)
		cmd.Env = append(cmd.Env, fmt.Sprintf("FAILING=%s", fail))
		return cmd
	}
}

func fakeCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestYarnHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestYarnHelperProcess(t *testing.T) {
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
	case "yarn":
		if os.Getenv("FAILING") == "yarn" {
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "non-yarn command called: cmd=%v args=%v", cmd, args)
		os.Exit(1)
	}

	switch args[0] {
	case "global":
		switch args[1] {
		case "dir":
			if os.Getenv("FAILING") == "dir" {
				os.Exit(1)
			}

			yarnDir, err := ioutil.TempDir("", "global")
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to create tempdir: %v", err)
				os.Exit(1)
			}
			fmt.Println(yarnDir)
		case "upgrade":
			if os.Getenv("FAILING") == "upgrade" {
				os.Exit(1)
			}
		default:
			fmt.Fprintf(os.Stderr, "unexpected flag/command provided: %v", args)
			os.Exit(2)
		}
	case "":
	default:
		fmt.Fprintf(os.Stderr, "unexpected flag/command provided: %v", args)
		os.Exit(1)
	}

	if os.Getenv("FAILING") == "true" {
		os.Exit(3)
	}
}
