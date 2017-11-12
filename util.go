package main

import (
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

// RelPath returns the given file's path relative to the configuration file.
func RelPath(file string) string {
	rel, err := filepath.Rel(runConfig.Config.ParentDir, file)
	if err != nil {
		log.WithError(err).WithField("file", file).Debug("Unable to get relative path for file")
		return file
	}

	return rel
}

// Constructs the symlinks necessary to be able to write to the file
func CreateNecessaryDirectories(file string) error {
	dir := filepath.Dir(file)
	return os.MkdirAll(dir, os.ModePerm)
}
