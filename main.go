package main

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

var (
	configFile = kingpin.Arg("config", "Configuration file").Required().String()
	logLevel   = kingpin.Flag("log-level", "Log level").Short('l').Default("info").Enum("debug", "info", "warning", "error")
	dryRun     = kingpin.Flag("dryrun", "Just print what would have been done").Short('n').Bool()

	runConfig = RunConfig{}
)

func main() {
	kingpin.Parse()

	runConfig = RunConfig{
		DryRun:     *dryRun,
		ConfigFile: *configFile,
	}

	logLevel, _ := log.ParseLevel(*logLevel)
	log.SetLevel(logLevel)

	runConfig.Config = readConfigFile(*configFile)
	log.WithFields(log.Fields{
		"config": runConfig, "file": *configFile,
	}).Debug("Successfully parsed config")

	var hadError = createSymlinks()
	if hadError {
		os.Exit(1)
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

	contents, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.WithField("file", file).WithError(err).Fatal("Unable to read file")
	}

	err = yaml.Unmarshal(contents, parsed)
	if err != nil {
		log.WithField("file", filename).WithError(err).Fatal("Unable to parse file as yaml")
	}

	log.Debug("Parsed yaml succesfully")
	return filename
}

func readConfigFile(file string) Config {
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
		backend := Backend{}

		readYamlFile(file, &backend)
		config.Backends[name] = backend
	}
}

func createSymlinks() bool {
	hadError := false

	for to, from := range runConfig.Config.Symlinks {
		log.WithFields(log.Fields{
			"to":   to,
			"from": from,
		}).Info("Creating symlink")

		from = filepath.Join(runConfig.Config.ParentDir, from)
		to = filepath.Join(runConfig.Config.ParentDir, to)

		_, err := os.Stat(from)
		if err != nil {
			log.WithError(err).WithField("from", relPath(from)).Warning("No such file")
			hadError = true
			continue
		}

		if runConfig.DryRun {
			continue
		}

		err = createNecessaryDirectories(to)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"from": relPath(from),
				"to":   relPath(to),
			}).Warning("Unable to create necessary directories")
			hadError = true

			continue
		}

		err = os.Symlink(from, to)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"from": relPath(from),
				"to":   relPath(to),
			}).Warning("Unable to create symlink")
			hadError = true
		}
	}

	return hadError
}

func createNecessaryDirectories(file string) error {
	dir := filepath.Dir(file)
	return os.MkdirAll(dir, os.ModePerm)
}

func relPath(file string) string {
	rel, err := filepath.Rel(runConfig.Config.ParentDir, file)
	if err != nil {
		log.WithError(err).WithField("file", file).Debug("Unable to get relative path for file")
		return file
	}

	return rel
}
