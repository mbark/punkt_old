package util

import (
	"os"
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
