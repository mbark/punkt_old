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
)

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

var _ = Describe("Manager", func() {
	BeforeEach(func() {
		logrus.SetLevel(logrus.PanicLevel)
	})

	It("should always return at least the git and symlink managers", func() {
		config := conf.Config{Managers: make(map[string]map[string]string)}

		root := mgr.NewRootManager(config)
		all := root.All()

		Expect(len(all)).To(Equal(2))
	})

	It("should return an additional manager if configured", func() {
		mgrs := make(map[string]map[string]string)
		mgrs["foo"] = make(map[string]string)

		config := conf.Config{Managers: mgrs}

		root := mgr.NewRootManager(config)
		all := root.All()

		Expect(len(all)).To(Equal(3))
	})

	Context("Dump", func() {
		var mockMgr *mockManager
		var root *mgr.RootManager
		var config conf.Config

		BeforeEach(func() {
			name := "foo"

			mgrs := make(map[string]map[string]string)
			mgrs[name] = make(map[string]string)
			config = conf.Config{
				Managers: mgrs,
				Fs:       memfs.New(),
			}
			root = mgr.NewRootManager(config)

			mockMgr = new(mockManager)
			mockMgr.On("Name", mock.Anything).Return(name)
		})

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

	Context("Ensure", func() {})

	Context("Update", func() {})
})
