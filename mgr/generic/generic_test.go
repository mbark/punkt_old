package generic_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/mgr/generic"
	"github.com/mbark/punkt/run"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-billy.v4/memfs"
)

func TestMgr(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Generic Suite")
}

func mockRun(mgr *generic.Manager) {
	mgr.PrintOutputToUser = func(c *exec.Cmd) {}
	mgr.WithCapture = func(cmd *exec.Cmd) (*bytes.Buffer, error) {
		out, _ := run.CaptureOutput(cmd)
		err := run.Run(cmd)

		return out, err
	}
}

const name = "generic"

var _ = Describe("Generic Manager", func() {
	var config *conf.Config
	var mgr *generic.Manager
	var configFile string

	BeforeEach(func() {
		logrus.SetLevel(logrus.PanicLevel)

		managers := make(map[string]map[string]string)
		managers[name] = make(map[string]string)
		managers[name]["command"] = name

		config = &conf.Config{
			UserHome:   "/home",
			PunktHome:  "/home/.config/punkt",
			Dotfiles:   "/home/.dotfiles",
			Fs:         memfs.New(),
			WorkingDir: "/home",
			Command:    fakeCommand,
			Managers:   managers,
		}

		configFile = filepath.Join(config.PunktHome, name+".toml")
		mgr = generic.NewManager(*config, configFile, name)
		mockRun(mgr)
	})

	// TODO: test for having no command?

	It("should have the name generic", func() {
		Expect(mgr.Name()).To(Equal("generic"))
	})

	var _ = Context("Dump", func() {
		It("should default to using generic", func() {
			out, err := mgr.Dump()
			Expect(err).To(BeNil())
			Expect(out).To(Equal(name + " dump"))
		})

		It("should fail if the command fails", func() {
			config.Command = fakeWithEnvCommand("FAILING=true")
			mgr = generic.NewManager(*config, configFile, name)
			_, err := mgr.Dump()

			Expect(err).NotTo(BeNil())
		})

		It("should prefer using 'dump' over 'command'", func() {
			config.Managers[name]["dump"] = "foo"
			mgr = generic.NewManager(*config, configFile, name)
			mockRun(mgr)

			out, err := mgr.Dump()
			Expect(err).To(BeNil())
			Expect(out).To(Equal("foo"))
		})
	})

	var _ = Context("Update", func() {
		It("should succeed if the command does", func() {
			err := mgr.Update()
			Expect(err).To(BeNil())
		})

		It("should fail if the command fails", func() {
			config.Command = fakeWithEnvCommand("FAILING=true")
			mgr = generic.NewManager(*config, configFile, name)
			err := mgr.Update()

			Expect(err).NotTo(BeNil())
		})
	})

	var _ = Context("Ensure", func() {
		It("should succeed if the command does", func() {
			err := mgr.Ensure()
			Expect(err).To(BeNil())
		})

		It("should fail if the command fails", func() {
			config.Command = fakeWithEnvCommand("FAILING=true")
			mgr = generic.NewManager(*config, configFile, name)
			err := mgr.Ensure()

			Expect(err).NotTo(BeNil())
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
	cs := []string{"-test.run=TestGenericHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestGenericHelperProcess(t *testing.T) {
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
		fmt.Fprintf(os.Stderr, "no command\n")
		os.Exit(2)
	}

	cmd, args := args[0], args[1:]

	if cmd != "sh" || args[0] != "-c" {
		fmt.Fprintf(os.Stderr, "should always use sh -c, cmd: %v, args: %v\n", cmd, args)
		os.Exit(1)
	}

	cmd, args = args[1], args[2:]
	if len(args) > 0 {
		os.Exit(1)
	}

	fmt.Print(cmd)
	os.Exit(0)
}
