package conf_test

import (
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	"github.com/mbark/punkt/pkg/conf"
	"github.com/mbark/punkt/pkg/fs"
	"github.com/mbark/punkt/pkg/test"
)

func TestConf(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Conf Suite")
}

var _ = Describe("Manager", func() {
	var snapshot fs.Snapshot
	var savedConfig map[string]string
	var configFile string

	BeforeEach(func() {
		snapshot, _ = test.MockSetup()
		configFile = filepath.Join(snapshot.WorkingDir, "config.toml")

		savedConfig = make(map[string]string)
		savedConfig["logLevel"] = "warn"
		savedConfig["dotfiles"] = "/some/where"
		savedConfig["punktHome"] = "/punkt/.home"
		err := snapshot.SaveToml(savedConfig, configFile)
		Expect(err).To(BeNil())
	})

	It("should read the given config file", func() {
		config, err := conf.NewConfig(snapshot, configFile)

		Expect(config).NotTo(BeNil())
		Expect(err).To(BeNil())
		Expect(logrus.GetLevel()).To(Equal(logrus.WarnLevel))
		Expect(config.Dotfiles).To(Equal(savedConfig["dotfiles"]))
		Expect(config.PunktHome).To(Equal(savedConfig["punktHome"]))
	})

	It("should handle when a relative file is given", func() {
		relPath, err := filepath.Rel(snapshot.WorkingDir, configFile)
		Expect(err).To(BeNil())

		config, err := conf.NewConfig(snapshot, relPath)

		Expect(config).NotTo(BeNil())
		Expect(err).To(BeNil())
		Expect(logrus.GetLevel()).To(Equal(logrus.WarnLevel))
		Expect(config.Dotfiles).To(Equal(savedConfig["dotfiles"]))
		Expect(config.PunktHome).To(Equal(savedConfig["punktHome"]))
	})

	It("should set a default for loglevel", func() {
		savedConfig["logLevel"] = "mumbojumbo"
		err := snapshot.SaveToml(savedConfig, configFile)
		Expect(err).To(BeNil())

		_, err = conf.NewConfig(snapshot, configFile)
		Expect(logrus.GetLevel()).To(Equal(logrus.InfoLevel))
		Expect(err).To(BeNil())
	})

	It("should read the managers.toml file for manager configuration", func() {
		mgrs := make(map[string]map[string]string)
		mgrs["foo"] = make(map[string]string)
		mgrs["foo"]["command"] = "bar"

		err := snapshot.SaveToml(mgrs, filepath.Join(savedConfig["punktHome"], "managers.toml"))
		Expect(err).To(BeNil())

		config, err := conf.NewConfig(snapshot, configFile)
		Expect(config.Managers).To(Equal(mgrs))
		Expect(err).To(BeNil())
	})
})
