package symlink_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-billy.v4/memfs"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/file"
	"github.com/mbark/punkt/mgr/symlink"
)

func TestSymlink(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Symlink Suite")
}

var _ = Describe("Symlink: Manager", func() {
	var config *conf.Config
	var mgr *symlink.Manager
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
		configFile = filepath.Join(config.PunktHome, "symlinks.toml")

		mgr = symlink.NewManager(*config, configFile)
	})

	It("should do nothing and always succeed with 'dump'", func() {
		config.Fs = nil
		_, err := mgr.Dump()
		Expect(err).To(Succeed())
	})

	It("should do nothing and always succeed with 'update'", func() {
		config.Fs = nil
		Expect(mgr.Update()).To(Succeed())
	})

	It("should succeed if there is an empty config file", func() {
		err := file.SaveToml(config.Fs, symlink.Config{}, configFile)
		Expect(err).To(BeNil())
		Expect(mgr.Ensure()).To(Succeed())
	})

	It("should add a symlink if the config file has it", func() {
		s := createFile(config, "file")
		err := file.SaveToml(config.Fs, symlink.Config{Symlinks: []symlink.Symlink{*s}}, configFile)
		Expect(err).To(BeNil())

		Expect(mgr.Ensure()).To(Succeed())
		Expect(s.Exists(*config)).To(BeTrue())
	})

	It("should try to create all symlinks even if some fail", func() {
		failing := symlink.Symlink{
			Target: "",
			Link:   "",
		}
		success := createFile(config, "afile")

		initial := symlink.Config{Symlinks: []symlink.Symlink{failing, *success}}
		err := file.SaveToml(config.Fs, initial, configFile)
		Expect(err).To(BeNil())

		Expect(mgr.Ensure()).NotTo(Succeed())
		Expect(failing.Exists(*config)).NotTo(BeTrue())
		Expect(success.Exists(*config)).To(BeTrue())
	})

	It("should succeed when a symlink already exists", func() {
		path := filepath.Join(config.UserHome, "file")
		_, err := config.Fs.Create(path)
		Expect(err).To(BeNil())
		s, err := mgr.Add(path, "")
		Expect(s.Ensure(*config)).To(Succeed())
		Expect(err).To(BeNil())
		Expect(s.Exists(*config)).To(BeTrue())

		err = file.SaveToml(config.Fs, symlink.Config{Symlinks: []symlink.Symlink{*s}}, configFile)
		Expect(err).To(BeNil())

		Expect(mgr.Ensure()).To(Succeed())
		Expect(s.Exists(*config)).To(BeTrue())
	})
})

func fakeCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestAddHelperProcess", "--", command}
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestAddHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	os.Exit(1)
}

func createFile(config *conf.Config, target string) *symlink.Symlink {
	target = filepath.Join(config.UserHome, target)
	_, err := config.Fs.Create(target)
	Expect(err).To(BeNil())

	return symlink.NewSymlink(*config, "", target)
}
