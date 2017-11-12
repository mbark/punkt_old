package main

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

// RunConfig contains the configuration for running, primarily set by the
// command line arguments
type RunConfig struct {
	DryRun     bool
	ConfigFile string
	Config     Config
}

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

		log.WithFields(log.Fields{
			"backend": name,
			"config":  backend,
		}).Debug("Parsed backend config")
	}
}

func readYamlFile(file string, parsed interface{}) string {
	filename, err := filepath.Abs(file)
	if err != nil {
		log.WithField("file", filename).WithError(err).Fatal("Unable to find absolute path to file")
	}
	_, err = os.Stat(filename)
	if os.IsNotExist(err) {
		log.WithField("file", filename).WithError(err).Fatal("Unable to find file")
	}

	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		log.WithField("file", file).WithError(err).Fatal("Unable to read file")
	}

	err = yaml.Unmarshal(contents, parsed)
	if err != nil {
		log.WithField("file", filename).WithError(err).Fatal("Unable to parse file as yaml")
	}

	log.WithFields(log.Fields{
		"content": string(contents),
		"file":    file,
	}).Debug("Parsed yaml succesfully")
	return filename
}
