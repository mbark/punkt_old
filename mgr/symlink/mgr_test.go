package symlink_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"gopkg.in/src-d/go-billy.v4/memfs"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/file"
	"github.com/mbark/punkt/mgr/symlink"
)

func TestSymlink(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Symlink Suite")
}

type mockLinkManager struct {
	mock.Mock
}

func (m *mockLinkManager) New(target, link string) *symlink.Symlink {
	args := m.Called(target, link)
	return args.Get(0).(*symlink.Symlink)
}

func (m *mockLinkManager) Remove(link string) (*symlink.Symlink, error) {
	args := m.Called(link)
	return args.Get(0).(*symlink.Symlink), args.Error(1)
}

func (m *mockLinkManager) Ensure(link *symlink.Symlink) error {
	args := m.Called(link)
	return args.Error(0)
}

func (m *mockLinkManager) Exists(link *symlink.Symlink) bool {
	args := m.Called(link)
	return args.Bool(0)
}

func (m *mockLinkManager) Expand(link symlink.Symlink) *symlink.Symlink {
	return &link
}

func (m *mockLinkManager) Unexpand(link symlink.Symlink) *symlink.Symlink {
	return &link
}

var _ = Describe("Symlink: Manager", func() {
	var config *conf.Config
	var linkMgr *mockLinkManager
	var mgr *symlink.Manager
	var configFile string

	var configWithLink = symlink.Config{
		Symlinks: []symlink.Symlink{
			{Target: "", Link: ""},
		},
	}

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
		linkMgr = new(mockLinkManager)
		mgr.LinkManager = linkMgr

		linkMgr.On("New", mock.Anything, mock.Anything).Return(new(symlink.Symlink))
		linkMgr.On("Ensure", mock.Anything).Return(nil)
	})

	It("should be called symlink", func() {
		Expect(mgr.Name()).To(Equal("symlink"))
	})

	var _ = Context("when running Dump", func() {
		It("should do nothing and always succeed", func() {
			out, err := mgr.Dump()
			Expect(err).To(BeNil())
			Expect(out).To(Equal(""))
		})
	})

	var _ = Context("when running update", func() {
		It("should do nothing and always succeed", func() {
			Expect(mgr.Update()).To(Succeed())
		})
	})

	var _ = Context("when running Ensure", func() {
		It("should succeed if there are no symlinks", func() {
			err := file.SaveToml(config.Fs, symlink.Config{}, configFile)
			Expect(err).To(BeNil())
			Expect(mgr.Ensure()).To(Succeed())
		})

		It("should succeed if the file can't be found", func() {
			Expect(mgr.Ensure()).To(Succeed())
		})

		It("should fail if the toml can't be read", func() {
			err := file.Save(config.Fs, "foo", configFile)
			Expect(err).To(BeNil())
			Expect(mgr.Ensure()).NotTo(Succeed())
		})

		It("should succeed if all symlinks can be ensured", func() {
			err := file.SaveToml(config.Fs, configWithLink, configFile)
			Expect(err).To(BeNil())
			Expect(mgr.Ensure()).To(Succeed())
		})

		It("should fail if some repo can't be ensured", func() {
			linkMgr = new(mockLinkManager)
			mgr.LinkManager = linkMgr
			linkMgr.On("New", mock.Anything, mock.Anything).Return(new(symlink.Symlink))
			linkMgr.On("Ensure", mock.Anything).Return(fmt.Errorf("fail"))

			err := file.SaveToml(config.Fs, configWithLink, configFile)
			Expect(err).To(BeNil())

			Expect(mgr.Ensure()).NotTo(Succeed())
		})
	})

	var _ = Context("when running Add", func() {
		// TODO: this test does nothing
		It("should make the target path absolute", func() {
			target := "relative"
			expected := filepath.Join(config.WorkingDir, target)
			linkMgr.On("New", mock.Anything, expected).Return(new(symlink.Symlink))

			_, err := mgr.Add(target, "/foo/bar")
			Expect(err).To(BeNil())
		})

		// TODO: this test does nothing
		It("should ensure the symlink exists", func() {
			_, err := mgr.Add("/a/file", "/some/where")
			Expect(err).To(BeNil())
		})

		It("should save the symlink added", func() {
			s, err := mgr.Add("/a/file", "/some/where")
			Expect(err).To(BeNil())

			var c symlink.Config
			err = file.ReadToml(config.Fs, &c, configFile)
			Expect(err).To(BeNil())

			Expect(c.Symlinks).To(ConsistOf(*s))
		})

		It("should not save the symlink if it already exists", func() {
			_, err := mgr.Add("/a/file", "/some/where")
			Expect(err).To(BeNil())
			s, err := mgr.Add("/a/file", "/some/where")
			Expect(err).To(BeNil())

			var c symlink.Config
			err = file.ReadToml(config.Fs, &c, configFile)
			Expect(err).To(BeNil())

			Expect(c.Symlinks).To(ConsistOf(*s))
		})
	})

	var _ = Context("when running Remove", func() {
		It("should succeed when removing a link that was added", func() {
			s, err := mgr.Add("/a/file", "/some/where")
			Expect(err).To(BeNil())
			linkMgr.On("Remove", mock.Anything).Return(s, nil)

			err = mgr.Remove(s.Link)
			Expect(err).To(BeNil())

			var c symlink.Config
			err = file.ReadToml(config.Fs, &c, configFile)
			Expect(err).To(BeNil())

			Expect(c.Symlinks).To(BeEmpty())
		})

		It("should succeed even if the symlink isn't stored in the config", func() {
			linkMgr.On("Remove", mock.Anything).Return(new(symlink.Symlink), nil)
			Expect(mgr.Remove("link")).To(Succeed())
		})

		It("should fail and not remove the link if it can't remove it", func() {
			linkMgr.On("Remove", mock.Anything).Return(new(symlink.Symlink), fmt.Errorf("fail"))
			s, err := mgr.Add("/a/file", "/some/where")
			Expect(err).To(BeNil())

			Expect(mgr.Remove(s.Link)).NotTo(Succeed())
		})
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
