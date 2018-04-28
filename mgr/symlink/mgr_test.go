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
	"gopkg.in/src-d/go-billy.v4/memfs"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/file"
	"github.com/mbark/punkt/mgr/symlink"
)

func TestSymlink(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Symlink Suite")
}

type fakeLinkManager struct {
	newer   func(string, string) *symlink.Symlink
	ensurer func(*symlink.Symlink) error
	exister func(*symlink.Symlink) bool
	remover func(string) (*symlink.Symlink, error)
}

func (mgr fakeLinkManager) New(target, link string) *symlink.Symlink {
	if mgr.newer != nil {
		return mgr.newer(target, link)
	}

	return &symlink.Symlink{Target: target, Link: link}
}

func (mgr fakeLinkManager) Remove(link string) (*symlink.Symlink, error) {
	if mgr.remover != nil {
		return mgr.remover(link)
	}

	return &symlink.Symlink{Target: "", Link: link}, nil
}

func (mgr fakeLinkManager) Ensure(link *symlink.Symlink) error {
	if mgr.ensurer != nil {
		return mgr.ensurer(link)
	}

	return nil
}

func (mgr fakeLinkManager) Exists(link *symlink.Symlink) bool {
	if mgr.exister != nil {
		return mgr.exister(link)
	}

	return false
}

func (mgr fakeLinkManager) Expand(link symlink.Symlink) *symlink.Symlink {
	return &link
}

func (mgr fakeLinkManager) Unexpand(link symlink.Symlink) *symlink.Symlink {
	return &link
}

var _ = Describe("Symlink: Manager", func() {
	var config *conf.Config
	var linkMgr *fakeLinkManager
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
		linkMgr = &fakeLinkManager{}
		mgr.LinkManager = linkMgr
	})

	It("should be called symlin", func() {
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
			linkMgr.ensurer = func(link *symlink.Symlink) error {
				return fmt.Errorf("fail")
			}

			err := file.SaveToml(config.Fs, configWithLink, configFile)
			Expect(err).To(BeNil())
			Expect(mgr.Ensure()).NotTo(Succeed())
		})
	})

	var _ = Context("when running Symlinks", func() {
		It("should return nothing if the config file doesn't exist", func() {
			err := file.SaveToml(config.Fs, symlink.Config{}, configFile)
			Expect(err).To(BeNil())
			Expect(mgr.Symlinks()).To(BeEmpty())
		})

		It("should return the stored symlinks", func() {
			err := file.SaveToml(config.Fs, configWithLink, configFile)
			Expect(err).To(BeNil())
			Expect(mgr.Symlinks()).To(Equal(configWithLink.Symlinks))
		})
	})

	var _ = Context("when running Add", func() {
		It("should handle when both target and new location are given", func() {
			target := "/a/file"
			location := "/some/where"

			linkMgr.newer = func(actualLocation, actualTarget string) *symlink.Symlink {
				Expect(actualLocation).To(Equal(location))
				Expect(actualTarget).To(Equal(target))

				return &symlink.Symlink{Link: actualLocation, Target: actualTarget}
			}

			_, err := mgr.Add(target, location)
			Expect(err).To(BeNil())
		})

		It("should make the target path absolute", func() {
			target := "relative"

			linkMgr.newer = func(_, actualTarget string) *symlink.Symlink {
				Expect(actualTarget).To(Equal(filepath.Join(config.WorkingDir, target)))

				return &symlink.Symlink{Link: "", Target: actualTarget}
			}

			_, err := mgr.Add(target, "/foo/bar")
			Expect(err).To(BeNil())
		})

		It("should derive the new location if none is given", func() {
			target := "/home/file"
			expected := filepath.Join(config.Dotfiles, "file")

			s, err := mgr.Add(target, "")

			Expect(err).To(BeNil())
			Expect(s.Target).To(Equal(expected))
		})

		It("should ensure the symlink exists", func() {
			ensured := false

			linkMgr.ensurer = func(link *symlink.Symlink) error {
				ensured = true
				return nil
			}

			_, err := mgr.Add("/a/file", "/some/where")
			Expect(err).To(BeNil())
			Expect(ensured).To(BeTrue())
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

			linkMgr.remover = func(link string) (*symlink.Symlink, error) {
				return s, nil
			}

			err = mgr.Remove(s.Link)
			Expect(err).To(BeNil())

			var c symlink.Config
			err = file.ReadToml(config.Fs, &c, configFile)
			Expect(err).To(BeNil())

			Expect(c.Symlinks).To(BeEmpty())
		})

		It("should succeed even if the symlink isn't stored in the config", func() {
			Expect(mgr.Remove("link")).To(Succeed())
		})

		It("should fail and not remove the link if it can't remove it", func() {
			s, err := mgr.Add("/a/file", "/some/where")
			Expect(err).To(BeNil())

			linkMgr.remover = func(link string) (*symlink.Symlink, error) {
				return nil, fmt.Errorf("fail")
			}

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
