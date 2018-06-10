package path

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/mbark/punkt/printer"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-billy.v4"
)

// CreateNecessaryDirectories constructs the directories necessary to be able to
// write to the file
func CreateNecessaryDirectories(fs billy.Filesystem, file string) error {
	dir := filepath.Dir(file)
	logrus.WithField("dir", dir).Debug("creating required directories")

	err := fs.MkdirAll(dir, os.ModePerm)
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
func AsAbsolute(fs billy.Filesystem, workingDir, file string) (string, error) {
	abs := file
	if !filepath.IsAbs(file) {
		abs = fs.Join(workingDir, file)
	}

	_, err := fs.Stat(abs)
	return abs, err
}

// GetUserHome returns the user's home directory
func GetUserHome() string {
	usr, err := user.Current()
	if err != nil {
		logrus.WithError(err).Fatal("unable to get user home")
		return ""
	}

	return usr.HomeDir
}

// ExpandHome takes the given string and replaces occurrences of ~ with the
// current user's home directory
func ExpandHome(s string, home string) string {
	return strings.Replace(s, "~", home, 1)
}

// UnexpandHome takes the given string and replaces the user's home with ~
// This is useful when you want to make something home-relative, rather than
// absolute
func UnexpandHome(s string, home string) string {
	return strings.Replace(s, home, "~", 1)
}
