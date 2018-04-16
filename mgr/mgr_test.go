package mgr_test

import (
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-billy.v4/memfs"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/mgr"
)

func TestMgr(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mgr Suite")
}

var _ = Describe("Manager", func() {
	BeforeEach(func() {
		logrus.SetLevel(logrus.PanicLevel)
	})

	It("should always return at least the git and symlink managers", func() {
		config := conf.Config{
			PunktHome:  "",
			Dotfiles:   "",
			UserHome:   "",
			WorkingDir: "",
			Fs:         memfs.New(),
			Command:    exec.Command,
			Managers:   make(map[string]map[string]string),
		}

		all := mgr.All(config)

		Expect(len(all)).To(Equal(2))
	})

	It("should return an additional manager if configured", func() {
		mgrs := make(map[string]map[string]string)
		mgrs["foo"] = make(map[string]string)

		config := conf.Config{
			PunktHome:  "",
			Dotfiles:   "",
			UserHome:   "",
			WorkingDir: "",
			Fs:         memfs.New(),
			Command:    exec.Command,
			Managers:   mgrs,
		}

		all := mgr.All(config)

		Expect(len(all)).To(Equal(3))
	})
})
