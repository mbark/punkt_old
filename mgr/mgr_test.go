package mgr_test

import (
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/src-d/go-billy.v4/memfs"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/mgr"
)

func TestMgr(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mgr Suite")
}

var _ = Describe("Manager", func() {
	It("should return an empty list if no managers are given", func() {
		config := conf.Config{
			PunktHome:  "",
			Dotfiles:   "",
			UserHome:   "",
			WorkingDir: "",
			Fs:         memfs.New(),
			Command:    exec.Command,
			Managers:   make(map[string]map[string]string),
		}

		Expect(len(mgr.All(config))).To(Equal(0))
	})

	It("should return a non-zero list if a manager is provided", func() {
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

		Expect(len(mgr.All(config))).To(Equal(1))
	})
})
