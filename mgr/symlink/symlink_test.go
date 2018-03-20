package symlink_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	g "github.com/onsi/ginkgo"
	m "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-billy.v4/memfs"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/file"
	"github.com/mbark/punkt/mgr/symlink"
)

func TestSymlink(t *testing.T) {
	m.RegisterFailHandler(g.Fail)
	g.RunSpecs(t, "Symlink Suite")
}

var _ = g.Describe("Symlink: Manager", func() {
	var config *conf.Config
	var mgr *symlink.Manager

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

		mgr = symlink.NewManager(*config)
	})

	g.It("should do nothing and always succeed with 'dump'", func() {
		config.Fs = nil
		m.Expect(mgr.Dump()).To(m.Succeed())
	})

	g.It("should do nothing and always succeed with 'update'", func() {
		config.Fs = nil
		m.Expect(mgr.Update()).To(m.Succeed())
	})

	g.It("should succeed if there is an empty config file", func() {
		err := file.SaveYaml(config.Fs, []symlink.Symlink{}, config.Dotfiles, "symlinks")
		m.Expect(err).To(m.BeNil())
		m.Expect(mgr.Ensure()).To(m.Succeed())
	})

	g.It("should fail if the symlink file can't be parsed", func() {
		err := file.Save(config.Fs, "foo", config.Dotfiles, "symlinks.yml")
		m.Expect(err).To(m.BeNil())
		m.Expect(mgr.Ensure()).NotTo(m.Succeed())
	})

	g.It("should add a symlink if the config file has it", func() {
		s := createFile(config, "/file", "/another/file")
		err := file.SaveYaml(config.Fs, []symlink.Symlink{*s}, config.Dotfiles, "symlinks")
		m.Expect(err).To(m.BeNil())

		m.Expect(mgr.Ensure()).To(m.Succeed())
		m.Expect(s.Exists()).To(m.BeTrue())
	})

	g.It("should try to create all symlinks even if some fail", func() {
		to := "/another/file"
		_, err := config.Fs.Create(config.UserHome + to)
		m.Expect(err).To(m.BeNil())

		fail1 := createFile(config, "/file", to)
		fail2 := symlink.NewSymlink(config.Fs, "/non/existant", "/some/where")
		success := createFile(config, "/afile", "/some/where")

		initial := []symlink.Symlink{*fail1, *fail2, *success}
		err = file.SaveYaml(config.Fs, initial, config.Dotfiles, "symlinks")
		m.Expect(err).To(m.BeNil())

		m.Expect(mgr.Ensure()).NotTo(m.Succeed())
		m.Expect(fail1.Exists()).NotTo(m.BeTrue())
		m.Expect(fail2.Exists()).NotTo(m.BeTrue())
		m.Expect(success.Exists()).To(m.BeTrue())
	})

	g.It("should succeed when a symlink already exists", func() {
		path := filepath.Join(config.UserHome, "file")
		_, err := config.Fs.Create(path)
		m.Expect(err).To(m.BeNil())
		s, err := mgr.Add(path)
		m.Expect(err).To(m.BeNil())
		m.Expect(s.Exists()).To(m.BeTrue())

		err = file.SaveYaml(config.Fs, []symlink.Symlink{*s}, config.Dotfiles, "symlinks")
		m.Expect(err).To(m.BeNil())

		m.Expect(mgr.Ensure()).To(m.Succeed())
		m.Expect(s.Exists()).To(m.BeTrue())
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

func createFile(config *conf.Config, from, to string) *symlink.Symlink {
	from = config.UserHome + from
	to = config.UserHome + to
	_, err := config.Fs.Create(from)
	m.Expect(err).To(m.BeNil())

	return symlink.NewSymlink(config.Fs, from, to)
}
