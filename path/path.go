package path

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// CreateNecessaryDirectories constructs the directories necessary to be able to
// write to the file
func CreateNecessaryDirectories(file string) error {
	dir := filepath.Dir(file)
	logrus.WithField("dir", dir).Debug("Creating required directories")
	return os.MkdirAll(dir, os.ModePerm)
}

// GoToPunktHome ...
func GoToPunktHome() {
	workingDir := GetPunktHome()

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

// GetUserHome returns the user's home directory
func GetUserHome() string {
	usr, err := user.Current()
	if err != nil {
		logrus.WithError(err).Fatal("Unable to get current user")
	}

	return usr.HomeDir
}

// GetPunktHome returns the directory for the punkt configuration
func GetPunktHome() string {
	return filepath.Join(GetUserHome(), ".config", "punkt")
}
