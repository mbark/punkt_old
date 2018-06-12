package generic_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/mbark/punkt/pkg/conf"
	"github.com/mbark/punkt/pkg/mgr/generic"
	"github.com/mbark/punkt/pkg/run"
	"github.com/mbark/punkt/pkg/run/runtest"
	"github.com/mbark/punkt/pkg/test"
)

func TestMgr(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Generic Suite")
}

const name = "generic"

var _ = Describe("Generic Manager", func() {
	var config conf.Config
	var mgr *generic.Manager
	var configFile string

	BeforeEach(func() {
		_, config = test.MockSetup()
		run.Commander = runtest.FakeCommand("TestGenericHelperProcess")

		managers := make(map[string]map[string]string)
		managers[name] = make(map[string]string)
		managers[name]["command"] = name
		config.Managers = managers

		configFile = filepath.Join(config.PunktHome, name+".toml")
		mgr = generic.NewManager(config, configFile, name)
	})

	It("should have the name generic", func() {
		Expect(mgr.Name()).To(Equal("generic"))
	})

	var _ = Context("Dump", func() {
		It("should default to using generic", func() {
			out, err := mgr.Dump()
			Expect(err).To(BeNil())
			Expect(out).To(Equal(name + " dump"))
		})

		It("should fail if the command fails", func() {
			run.Commander = runtest.FakeWithEnvCommand("TestGenericHelperProcess", "FAILING=true")
			mgr = generic.NewManager(config, configFile, name)
			_, err := mgr.Dump()

			Expect(err).NotTo(BeNil())
		})

		It("should prefer using 'dump' over 'command'", func() {
			config.Managers[name]["dump"] = "foo"
			mgr = generic.NewManager(config, configFile, name)

			out, err := mgr.Dump()
			Expect(err).To(BeNil())
			Expect(out).To(Equal("foo"))
		})
	})

	var _ = Context("Update", func() {
		It("should succeed if the command does", func() {
			err := mgr.Update()
			Expect(err).To(BeNil())
		})

		It("should fail if the command fails", func() {
			run.Commander = runtest.FakeWithEnvCommand("TestGenericHelperProcess", "FAILING=true")
			mgr = generic.NewManager(config, configFile, name)
			err := mgr.Update()

			Expect(err).NotTo(BeNil())
		})
	})

	var _ = Context("Ensure", func() {
		It("should succeed if the command does", func() {
			err := mgr.Ensure()
			Expect(err).To(BeNil())
		})

		It("should fail if the command fails", func() {
			run.Commander = runtest.FakeWithEnvCommand("TestGenericHelperProcess", "FAILING=true")
			mgr = generic.NewManager(config, configFile, name)
			err := mgr.Ensure()

			Expect(err).NotTo(BeNil())
		})
	})
})

func TestGenericHelperProcess(t *testing.T) {
	cmd, args, err := runtest.VerifyHelperProcess()
	if err != nil {
		return
	}

	if cmd != "sh" || args[0] != "-c" {
		fmt.Fprintf(os.Stderr, "should always use sh -c, cmd: %v, args: %v\n", cmd, args)
		os.Exit(1)
	}

	cmd, args = args[1], args[2:]
	if len(args) > 0 {
		os.Exit(1)
	}

	fmt.Print(cmd)
	os.Exit(0)
}
