package mgr_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/BurntSushi/toml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"gopkg.in/src-d/go-billy.v4/memfs"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/file"
	"github.com/mbark/punkt/mgr"
	"github.com/mbark/punkt/mgr/symlink"
	"github.com/mbark/punkt/mgr/symlink/symlinktest"
)

type managerConfig struct {
	Symlinks []symlink.Symlink
}

type mockManager struct {
	mock.Mock
}

func (m *mockManager) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockManager) Dump() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *mockManager) Ensure() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockManager) Update() error {
	args := m.Called()
	return args.Error(0)
}

func TestMgr(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mgr Suite")
}

const name = "foo"

var _ = Describe("Manager", func() {
	var mockMgr *mockManager
	var linkMgr *symlinktest.MockLinkManager
	var config conf.Config
	var root *mgr.RootManager

	BeforeEach(func() {
		logrus.SetLevel(logrus.PanicLevel)

		mgrs := make(map[string]map[string]string)
		mgrs[name] = make(map[string]string)
		config = conf.Config{
			Managers: mgrs,
			Fs:       memfs.New(),
		}
		root = mgr.NewRootManager(config)

		mockMgr = new(mockManager)
		mockMgr.On("Name", mock.Anything).Return(name)

		linkMgr = new(symlinktest.MockLinkManager)
		root.LinkManager = linkMgr
	})

	Context("All", func() {
		It("should always return at least the git and symlink managers", func() {
			config := conf.Config{Managers: make(map[string]map[string]string)}

			root := mgr.NewRootManager(config)
			all := root.All()

			Expect(len(all)).To(Equal(2))
		})

		It("should return an additional manager if configured", func() {
			all := root.All()

			Expect(len(all)).To(Equal(3))
		})
	})

	Context("Dump", func() {
		It("should succeed if all managers succeed and return empty string", func() {
			mockMgr.On("Dump", mock.Anything).Return("", nil)

			Expect(root.Dump([]mgr.Manager{mockMgr})).To(Succeed())
		})

		It("should fail if a manager fails", func() {
			mockMgr.On("Dump", mock.Anything).Return("", fmt.Errorf("fail"))

			Expect(root.Dump([]mgr.Manager{mockMgr})).NotTo(Succeed())
		})

		It("should save the dumped output to the config file", func() {
			expected := make(map[string]string)
			expected["foo"] = "bar"

			var out bytes.Buffer
			encoder := toml.NewEncoder(&out)
			Expect(encoder.Encode(expected)).To(Succeed())

			mockMgr.On("Dump", mock.Anything).Return(out.String(), nil)

			Expect(root.Dump([]mgr.Manager{mockMgr})).To(Succeed())

			var actual map[string]string
			err := file.ReadToml(config.Fs, &actual, root.ConfigFile("foo"))
			Expect(err).To(BeNil())

			Expect(actual).To(Equal(expected))
		})
	})

	Context("Ensure", func() {
		It("should succeed if the managers succeed and has no config files", func() {
			mockMgr.On("Ensure").Return(nil)

			Expect(root.Ensure([]mgr.Manager{mockMgr})).To(Succeed())
		})

		It("should fail if some manager fails", func() {
			mockMgr.On("Ensure").Return(fmt.Errorf("fail"))

			Expect(root.Ensure([]mgr.Manager{mockMgr})).NotTo(Succeed())
		})

		It("should ensure the symlink exists for the managers", func() {
			mockMgr.On("Ensure").Return(nil)
			linkMgr.On("Ensure", mock.Anything).Return(nil)
			expected := symlink.Symlink{
				Link:   "/link",
				Target: "/target",
			}

			mgrConfig := managerConfig{Symlinks: []symlink.Symlink{expected}}

			err := file.SaveToml(config.Fs, mgrConfig, root.ConfigFile(name))
			Expect(err).To(BeNil())

			Expect(root.Ensure([]mgr.Manager{mockMgr})).To(Succeed())

			linkMgr.AssertNumberOfCalls(GinkgoT(), "Ensure", 1)
		})

		It("should fail if some symlink doesn't exist", func() {
			mockMgr.On("Ensure").Return(nil)
			linkMgr.On("Ensure", mock.Anything).Return(fmt.Errorf("fail"))
			expected := symlink.Symlink{
				Link:   "/link",
				Target: "/target",
			}

			mgrConfig := managerConfig{Symlinks: []symlink.Symlink{expected}}

			err := file.SaveToml(config.Fs, mgrConfig, root.ConfigFile(name))
			Expect(err).To(BeNil())

			Expect(root.Ensure([]mgr.Manager{mockMgr})).NotTo(Succeed())
		})

		It("should fail if some config file can't be parsed", func() {
			mockMgr.On("Ensure").Return(nil)
			linkMgr.On("Ensure", mock.Anything).Return(nil)

			err := file.Save(config.Fs, "foo", root.ConfigFile(name))
			Expect(err).To(BeNil())

			Expect(root.Ensure([]mgr.Manager{mockMgr})).NotTo(Succeed())
		})

		It("should succeed even if the toml file doesn't contain a symlinks key", func() {
			mockMgr.On("Ensure").Return(nil)
			linkMgr.On("Ensure", mock.Anything).Return(nil)

			err := file.Save(config.Fs, "[foo]", root.ConfigFile(name))
			Expect(err).To(BeNil())

			Expect(root.Ensure([]mgr.Manager{mockMgr})).To(Succeed())
		})
	})

	Context("Update", func() {
		It("should succeed if all managers do", func() {
			mockMgr.On("Update").Return(nil)

			Expect(root.Update([]mgr.Manager{mockMgr})).To(Succeed())
		})

		It("should fail if a manager fails", func() {
			mockMgr.On("Update").Return(fmt.Errorf("fail"))

			Expect(root.Update([]mgr.Manager{mockMgr})).NotTo(Succeed())
		})
	})
})
