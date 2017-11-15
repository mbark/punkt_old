package backend

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mbark/punkt/opt"
	"github.com/mbark/punkt/path"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Backend contains the backend configuration for a specified backend
type Backend struct {
	List    string `yaml:"list"`
	Update  string `yaml:"update"`
	Install string `yaml:"install"`
}

// WriteInstalledPackages uses the configuration for the backend to get a list
// of all installed packages and write it to a file, called a database.
func (backend Backend) WriteInstalledPackages(name string, pkgdbs string) bool {
	listCmd := strings.Split(backend.List, " ")
	cmd := exec.Command(listCmd[0], listCmd[1:]...)
	out, err := cmd.Output()

	logger := logrus.WithFields(logrus.Fields{
		"backend": name,
		"cmd":     backend.List,
	})

	if err != nil {
		logger.WithError(err).Error("Unable to run command")
		return false
	}

	logger.Debug("Successfully listed installed packages")
	packages := strings.Split(string(out), "\n")

	return writeInstalledPackagesToFile(name, pkgdbs, packages)
}

// CreatePackageDirectory will create the necessary directories to be able to
// save the backend database files.
func CreatePackageDirectory(dir string) {
	if opt.DryRun {
		logrus.WithField("dir", dir).Info("Ensuring directories exist")
	}

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"dir": dir,
		}).WithError(err).Fatal("Unable to create directories")
	}
}

func writeInstalledPackagesToFile(name string, pkgdbs string, packages []string) bool {
	file := name + ".yaml"
	file = filepath.Join(pkgdbs, file)

	logger := logrus.WithFields(logrus.Fields{
		"file":    file,
		"backend": name,
	})

	err := path.CreateNecessaryDirectories(file)
	if err != nil {
		logger.WithError(err).Error("Unable to create necessary directories")
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
