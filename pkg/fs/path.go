package fs

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mbark/punkt/pkg/printer"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// CreateNecessaryDirectories constructs the directories necessary to be able to
// write to the file
func (snapshot Snapshot) CreateNecessaryDirectories(file string) error {
	dir := filepath.Dir(file)
	logrus.WithField("dir", dir).Debug("creating required directories")

	err := snapshot.Fs.MkdirAll(dir, os.ModePerm)
	if err != nil {
		printer.Log.Error("unable to create directories %s", dir)
		logrus.WithField("dir", dir).WithError(err).Error("unable to create necessary directories")
		return errors.Wrapf(err, "unable to create directories")
	}

	return nil
}

// AsAbsolute makes sure the given file is transformed to an absolute path, it
// also checks if the given file exists -- otherwiser returning an error. If this
// check isn't relevant it can just be ignored: the given path is still absolute.
func (snapshot Snapshot) AsAbsolute(file string) (string, error) {
	abs := file
	if !filepath.IsAbs(file) {
		abs = snapshot.Fs.Join(snapshot.WorkingDir, file)
	}

	_, err := snapshot.Fs.Stat(abs)
	return abs, err
}

// ExpandHome takes the given string and replaces occurrences of ~ with the
// current user's home directory
func (snapshot Snapshot) ExpandHome(s string) string {
	return strings.Replace(s, "~", snapshot.UserHome, 1)
}

// UnexpandHome takes the given string and replaces the user's home with ~
// This is useful when you want to make something home-relative, rather than
// absolute
func (snapshot Snapshot) UnexpandHome(s string) string {
	return strings.Replace(s, snapshot.UserHome, "~", 1)
}
