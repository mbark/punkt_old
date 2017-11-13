package backends

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mbark/punkt/config"
	"github.com/mbark/punkt/util"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// WriteInstalledPackages uses the configuration for the backend to get a list
// of all installed packages and write it to a file, called a database.
func WriteInstalledPackages(conf config.Config, backend config.Backend) bool {
	listCmd := strings.Split(backend.List, " ")
	cmd := exec.Command(listCmd[0], listCmd[1:]...)
	out, err := cmd.Output()

	logger := logrus.WithFields(logrus.Fields{
		"backend": backend.Name,
		"cmd":     backend.List,
	})

	if err != nil {
		logger.WithError(err).Error("Unable to run command")
		return false
	}

	logger.Debug("Successfully listed installed packages")

	packages := strings.Split(string(out), "\n")
	return writeInstalledPackagesToFile(conf, backend, packages)
}

// CreatePackageDirectory will create the necessary directories to be able to
// save the backend database files.
func CreatePackageDirectory(conf config.Config, dryRun bool) {
	dir := conf.PackageFiles
	err := os.MkdirAll(dir, os.ModePerm)

	if dryRun {
		logrus.WithField("dir", dir).Info("Ensuring directories exist")
	}

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"dir": dir,
		}).Fatal("Unable to create directories")
	}
}

func writeInstalledPackagesToFile(conf config.Config, backend config.Backend, packages []string) bool {
	file := backend.Name + ".yaml"
	file = filepath.Join(conf.ParentDir, conf.PackageFiles, file)

	logger := logrus.WithFields(logrus.Fields{
		"file":    conf.RelPath(file),
		"backend": backend.Name,
	})

	err := util.CreateNecessaryDirectories(file)
	if err != nil {
		logger.Error("Unable to create necessary directories")
		return false
	}

	f, err := os.Create(file)
	if err != nil {
		logger.WithError(err).Error("Unable to create file")
		return false
	}

	defer f.Close()

	out, err := yaml.Marshal(packages)
	if err != nil {
		logrus.WithError(err).Error("Unable to marshal packages to yaml")
		return false
	}

	_, err = f.Write(out)
	if err != nil {
		logger.WithError(err).Error("Unable to write to file")
		return false
	}

	f.Sync()

	logger.Info("Succesfully wrote to backend database file")
	return true
}
