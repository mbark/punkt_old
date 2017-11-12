package main

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// WriteInstalledPackages uses the configuration for the backend to get a list
// of all installed packages and write it to a file, called a database.
func WriteInstalledPackages(backend Backend) bool {
	listCmd := strings.Split(backend.List, " ")
	cmd := exec.Command(listCmd[0], listCmd[1:]...)
	out, err := cmd.Output()

	if err != nil {
		log.WithFields(log.Fields{
			"backend": backend.Name,
			"cmd":     backend.List,
		}).WithError(err).Error("Unable to run command")
		return false
	}

	log.WithFields(log.Fields{
		"backend": backend.Name,
		"cmd":     backend.List,
	}).Debug("Successfully listed installed packages")

	packages := strings.Split(string(out), "\n")
	return writeInstalledPackagesToFile(backend, packages)
}

// CreatePackageDirectory will create the necessary directories to be able to
// save the backend database files.
func CreatePackageDirectory() {
	dir := runConfig.Config.PackageFiles
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.WithFields(log.Fields{
			"directory": dir,
		}).Fatal("Unable to create directories")
	}
}

func writeInstalledPackagesToFile(backend Backend, packages []string) bool {
	file := backend.Name + ".yaml"
	file = filepath.Join(runConfig.Config.ParentDir, runConfig.Config.PackageFiles, file)

	err := CreateNecessaryDirectories(file)
	if err != nil {
		log.WithFields(log.Fields{
			"file": RelPath(file),
		}).Error("Unable to create necessary directories")
		return false
	}

	f, err := os.Create(file)
	if err != nil {
		log.WithFields(log.Fields{
			"file": file,
		}).Error("Unable to create file")
		return false
	}

	defer f.Close()

	out, err := yaml.Marshal(packages)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"backend": backend.Name,
		}).Error("Unable to marshal packages to yaml")
		return false
	}

	log.WithFields(log.Fields{
		"out":      string(out),
		"packages": packages,
	}).Info("Yaml to save")

	_, err = f.Write(out)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"file": file,
		}).Error("Unable to write to file")
		return false
	}

	f.Sync()

	log.WithFields(log.Fields{
		"backend": backend.Name,
		"file":    RelPath(file),
	}).Info("Succesfully wrote to backend database file")
	return true
}
