package path

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// CreateNecessaryDirectories constructs the symlinks necessary to be able to
// write to the file
func CreateNecessaryDirectories(file string) error {
	dir := filepath.Dir(file)
	logrus.WithField("dir", dir).Info("Creating required directories")
	return os.MkdirAll(dir, os.ModePerm)
}

// GoToPunktHome ...
func GoToPunktHome() {
	usr, _ := user.Current()
	dir := usr.HomeDir
	workingDir := filepath.Join(dir, ".config", "punkt")

	if err := os.MkdirAll(workingDir, os.ModePerm); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"workingdir": workingDir,
		}).Fatal("Could not create configuration home for punkt")
	}

	if err := os.Chdir(workingDir); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"workingdir": workingDir,
		}).Fatal("Unable to change working directory to that of the config file")
	}
}
