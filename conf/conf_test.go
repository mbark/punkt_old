package conf_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	billy "gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/osfs"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/file"
)

func TestConf(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Conf Suite")
}

var _ = Describe("Manager", func() {
	BeforeEach(func() {
		logrus.SetLevel(logrus.PanicLevel)
	})

	It("should use defaults if the config file doesn't exist", func() {
		config := conf.NewConfig("")
		Expect(config).NotTo(BeNil())
		Expect(logrus.GetLevel()).To(Equal(logrus.InfoLevel))
	})

	Context("with an existing config file", func() {
		var dir string
		var home string
		var savedConfig map[string]string
		var configFile string
		var fs billy.Filesystem

		BeforeEach(func() {
			d, err := ioutil.TempDir("", "conf")
			Expect(err).To(BeNil())

			dir = d

			home, err = ioutil.TempDir("", "conf")
			Expect(err).To(BeNil())

			fs = osfs.New("/")

			savedConfig = make(map[string]string)
			savedConfig["logLevel"] = "warn"
			savedConfig["dotfiles"] = "/some/where"
			savedConfig["punktHome"] = home
			err = file.SaveToml(fs, savedConfig, filepath.Join(dir, "config.toml"))
			Expect(err).To(BeNil())

			configFile = filepath.Join(dir, "config.toml")
		})

		AfterEach(func() {
			Expect(os.RemoveAll(dir)).To(Succeed())
			Expect(os.RemoveAll(home)).To(Succeed())
		})

		It("should read the given config file", func() {
			config := conf.NewConfig(configFile)

			Expect(config).NotTo(BeNil())
			Expect(logrus.GetLevel()).To(Equal(logrus.WarnLevel))
			Expect(config.Dotfiles).To(Equal(savedConfig["dotfiles"]))
			Expect(config.PunktHome).To(Equal(savedConfig["punktHome"]))
		})

		It("should handle when a relative file is given", func() {
			wd, err := os.Getwd()
			Expect(err).To(BeNil())
			relPath, err := filepath.Rel(wd, configFile)
			Expect(err).To(BeNil())

			config := conf.NewConfig(relPath)

			Expect(config).NotTo(BeNil())
			Expect(logrus.GetLevel()).To(Equal(logrus.WarnLevel))
			Expect(config.Dotfiles).To(Equal(savedConfig["dotfiles"]))
			Expect(config.PunktHome).To(Equal(savedConfig["punktHome"]))
		})

		It("should set a default for loglevel", func() {
			savedConfig["logLevel"] = "mumbojumbo"
			err := file.SaveToml(fs, savedConfig, filepath.Join(dir, "config.toml"))
			Expect(err).To(BeNil())

			conf.NewConfig(configFile)
			Expect(logrus.GetLevel()).To(Equal(logrus.InfoLevel))
		})

		It("should read the managers.toml file for manager configuration", func() {
			logrus.SetLevel(logrus.DebugLevel)
			mgrs := make(map[string]map[string]string)
			mgrs["foo"] = make(map[string]string)
			mgrs["foo"]["command"] = "bar"

			err := file.SaveToml(fs, mgrs, filepath.Join(home, "managers.toml"))
			Expect(err).To(BeNil())

			config := conf.NewConfig(configFile)
			Expect(config.Managers).To(Equal(mgrs))
		})
	})
})
