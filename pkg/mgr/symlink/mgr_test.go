package symlink_test

import (
	"fmt"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	"github.com/mbark/punkt/pkg/conf"
	"github.com/mbark/punkt/pkg/fs"
	"github.com/mbark/punkt/pkg/mgr/symlink"
	"github.com/mbark/punkt/pkg/mgr/symlink/symlinktest"
	"github.com/mbark/punkt/pkg/test"
)

func TestSymlink(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Symlink Suite")
}

var _ = Describe("Symlink: Manager", func() {
	var snapshot fs.Snapshot
	var config conf.Config
	var linkMgr *symlinktest.MockLinkManager
	var mgr *symlink.Manager
	var configFile string
	var existingFile string

	BeforeEach(func() {
		snapshot, config = test.MockSetup()

		existingFile = filepath.Join(snapshot.WorkingDir, ".configFile")
		_, err := snapshot.Fs.Create(existingFile)
		Expect(err).To(BeNil())

		configFile = filepath.Join(config.PunktHome, "symlinks.toml")

		mgr = symlink.NewManager(config, snapshot, configFile)
		linkMgr = new(symlinktest.MockLinkManager)
		mgr.LinkManager = linkMgr

		linkMgr.On("New", mock.Anything, mock.Anything).Return(&symlink.Symlink{
			Target: "target",
			Link:   "link",
		})
		linkMgr.On("Ensure", mock.Anything).Return(nil)

	})

	It("should be called symlink", func() {
		Expect(mgr.Name()).To(Equal("symlink"))
	})

	var _ = Context("Dump", func() {
		It("should do nothing and always succeed", func() {
			out, err := mgr.Dump()
			Expect(err).To(BeNil())
			Expect(out).To(Equal(""))
		})
	})

	var _ = Context("Update", func() {
		It("should do nothing and always succeed", func() {
			Expect(mgr.Update()).To(Succeed())
		})
	})

	var _ = Context("Dump", func() {
		It("should do nothing and always succeed", func() {
			out, err := mgr.Dump()
			Expect(out).To(Equal(""))
			Expect(err).To(BeNil())
		})
	})

	var _ = Context("Add", func() {
		It("should make the target path absolute", func() {
			target := filepath.Base(existingFile)
			location := "/foo/bar"
			expected := filepath.Join(snapshot.WorkingDir, target)

			_, err := mgr.Add(target, location)
			Expect(err).To(BeNil())

			linkMgr.AssertCalled(GinkgoT(), "New", location, expected)
		})

		It("should ensure the symlink exists", func() {
			linkMgr = new(symlinktest.MockLinkManager)
			mgr.LinkManager = linkMgr
			linkMgr.On("New", mock.Anything, mock.Anything).Return(&symlink.Symlink{
				Target: "target",
				Link:   "link",
			})
			linkMgr.On("Ensure", mock.Anything).Return(nil)

			_, err := mgr.Add(existingFile, "/some/where")
			Expect(err).To(BeNil())

			linkMgr.AssertCalled(GinkgoT(), "Ensure", mock.Anything)
		})

		It("should fail if the symlink can't be ensured", func() {
			linkMgr = new(symlinktest.MockLinkManager)
			mgr.LinkManager = linkMgr
			linkMgr.On("New", mock.Anything, mock.Anything).Return(&symlink.Symlink{
				Target: "target",
				Link:   "link",
			})
			linkMgr.On("Ensure", mock.Anything).Return(fmt.Errorf("fail"))

			_, err := mgr.Add(existingFile, "/location")
			Expect(err).NotTo(BeNil())
		})

		It("should save the symlink added", func() {
			s, err := mgr.Add(existingFile, "/some/where")
			Expect(err).To(BeNil())

			var c symlink.Config
			err = snapshot.ReadToml(&c, configFile)
			Expect(err).To(BeNil())

			Expect(c.Symlinks).To(ConsistOf(*s))
		})

		It("should not save the symlink if it already exists", func() {
			_, err := mgr.Add(existingFile, "/some/where")
			Expect(err).To(BeNil())
			s, err := mgr.Add(existingFile, "/some/where")
			Expect(err).To(BeNil())

			var c symlink.Config
			err = snapshot.ReadToml(&c, configFile)
			Expect(err).To(BeNil())

			Expect(c.Symlinks).To(ConsistOf(*s))
		})

		It("should fail if the stored config can't be parsed", func() {
			err := snapshot.Save("foo", configFile)
			Expect(err).To(BeNil())

			_, err = mgr.Add("/target", "/location")
			Expect(err).NotTo(BeNil())
		})

		It("should fail if the file to add doesn't exist", func() {
			_, err := mgr.Add("/a/file", "")
			Expect(err).NotTo(BeNil())
		})
	})

	var _ = Context("Remove", func() {
		It("should succeed when removing a link that was added", func() {
			s, err := mgr.Add(existingFile, "")
			Expect(err).To(BeNil())
			linkMgr.On("Remove", mock.Anything).Return(s, nil)

			err = mgr.Remove(existingFile)
			Expect(err).To(BeNil())

			var c symlink.Config
			err = snapshot.ReadToml(&c, configFile)
			Expect(err).To(BeNil())

			Expect(c.Symlinks).To(BeEmpty())
		})

		It("should succeed if the config file doesn't exist", func() {
			linkMgr.On("Remove", mock.Anything).Return(new(symlink.Symlink), nil)
			Expect(mgr.Remove(existingFile)).To(Succeed())
		})

		It("should succeed even if the symlink isn't stored in the config file", func() {
			linkMgr.On("Remove", mock.Anything).Return(new(symlink.Symlink), nil)
			_, err := mgr.Add(existingFile, "/some/where")
			Expect(err).To(BeNil())

			Expect(mgr.Remove(existingFile)).To(Succeed())
		})

		It("should fail and not remove the link if it can't remove it", func() {
			linkMgr.On("Remove", mock.Anything).Return(new(symlink.Symlink), fmt.Errorf("fail"))
			s, err := mgr.Add(existingFile, "/some/where")
			Expect(err).To(BeNil())

			Expect(mgr.Remove(s.Link)).NotTo(Succeed())
		})

		It("should handle relative paths", func() {
			linkMgr.On("Remove", mock.Anything).Return(new(symlink.Symlink), nil)
			mgr.Add(existingFile, "")

			relPath, err := filepath.Rel(snapshot.WorkingDir, existingFile)
			Expect(err).To(BeNil())

			Expect(mgr.Remove(relPath)).To(Succeed())

			linkMgr.AssertCalled(GinkgoT(), "Remove", existingFile)
		})

		It("should fail if the file doesn't exist", func() {
			Expect(mgr.Remove("/non/existant")).NotTo(Succeed())
		})
	})
})
