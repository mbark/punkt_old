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
	It("should return a non-zero length list", func() {
		config := conf.Config{
			PunktHome:  "",
			Dotfiles:   "",
			UserHome:   "",
			WorkingDir: "",
			Fs:         memfs.New(),
			Command:    exec.Command,
		}

		Expect(len(mgr.All(config))).NotTo(Equal(0))
	})
})
