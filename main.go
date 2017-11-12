package main

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

var (
	configFile = kingpin.Arg("config", "Configuration file").Required().String()
	logLevel   = kingpin.Flag("log-level", "Log level").Short('l').Default("debug").Enum("debug", "info", "warning", "error")
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

	runConfig.Config = ParseConfig(*configFile)
	log.WithFields(log.Fields{
		"config": runConfig,
		"file":   *configFile,
	}).Debug("Successfully parsed config")

	hadError := CreateSymlinks()
	CreatePackageDirectory()
	for _, val := range runConfig.Config.Backends {
		hadError = WriteInstalledPackages(val) && hadError
	}

	if hadError {
		os.Exit(1)
	}
}
