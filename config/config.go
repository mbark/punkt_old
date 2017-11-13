package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Config contains the user's configuration yaml file
type Config struct {
	ParentDir    string
	Symlinks     map[string]string   `yaml:"symlinks"`
	BackendFiles map[string]string   `yaml:"backends"`
	Backends     map[string]Backend  `yaml:"-"`
	Tasks        []map[string]string `yaml:"tasks"`
	PackageFiles string              `yaml:"package_files"`
}

// Backend contains the backend configuration for a specified backend
type Backend struct {
	Name    string `yaml:"-"`
	List    string `yaml:"list"`
	Update  string `yaml:"update"`
	Install string `yaml:"install"`
}

// RelPath returns the path relative to the configurations directory
func (config Config) RelPath(file string) string {
	rel, err := filepath.Rel(config.ParentDir, file)
	if err != nil {
		logrus.WithError(err).WithField("file", file).Debug("Unable to get relative path for file")
		return file
	}

	return rel
}

// ParseConfig parses the given configuration file and returns the Config struct
// Any errors at this stage will result in a fatal error.
func ParseConfig(file string) Config {
	config := Config{}
	filename := readYamlFile(file, &config)

	config.ParentDir = filepath.Dir(filename)
	readBackendConfigurations(&config)
	return config
}

func readBackendConfigurations(config *Config) {
	config.Backends = make(map[string]Backend)
	for name, file := range config.BackendFiles {
		file = filepath.Join(config.ParentDir, file)
		backend := Backend{
			Name: name,
		}

		readYamlFile(file, &backend)
		config.Backends[name] = backend

		logrus.WithFields(logrus.Fields{
			"backend": name,
			"config":  backend,
		}).Debug("Parsed backend config")
	}
}

func readYamlFile(file string, parsed interface{}) string {
	logger := logrus.WithFields(logrus.Fields{
		"file": file,
	})

	filename, err := filepath.Abs(file)
	if err != nil {
		logger.WithError(err).Fatal("Unable to find absolute path to file")
	}
	_, err = os.Stat(filename)
	if os.IsNotExist(err) {
		logger.WithError(err).Fatal("Unable to find file")
	}

	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		logger.WithError(err).Fatal("Unable to read file")
	}

	err = yaml.Unmarshal(contents, parsed)
	if err != nil {
		logger.WithError(err).Fatal("Unable to parse file as yaml")
	}

	logger.WithField("content", contents).Debug("Parsed yaml succesfully")
	return filename
}
