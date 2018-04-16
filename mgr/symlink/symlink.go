package symlink

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/path"
)

var (
	// ErrNonHomeRelativeTarget is returned if a target has no given link and is outside of the user's home directory
	ErrNonHomeRelativeTarget = errors.New("non-home relative target given without specific link location")
)

// Symlink describes a symlink, i.e. what it links from and what it links to
type Symlink struct {
	Target string
	Link   string
}

// NewSymlink creates a new symlink
func NewSymlink(config conf.Config, target, link string) *Symlink {
	logger := logrus.WithFields(logrus.Fields{
		"target": target,
		"link":   link,
	})
	logger.Debug("creating symlink")

	if target == "" && link != "" {
		logger.Debug("empty target with non-empty link, deriving target")
		target, _ = deriveLink(link, config.UserHome, config.Dotfiles)
	}

	if target != "" && link == "" {
		logger.Debug("empty link with non-empty target, deriving link")
		link, _ = deriveLink(target, config.Dotfiles, config.UserHome)
	}

	return &Symlink{
		Target: path.ExpandHome(target, config.UserHome),
		Link:   path.ExpandHome(link, config.UserHome),
	}
}

// Expand ...
func (symlink *Symlink) Expand(home string) {
	symlink.Target = path.ExpandHome(symlink.Target, home)
	symlink.Link = path.ExpandHome(symlink.Link, home)
}

// Unexpand ...
func (symlink *Symlink) Unexpand(home string) {
	symlink.Target = path.UnexpandHome(symlink.Target, home)
	symlink.Link = path.UnexpandHome(symlink.Link, home)
}

// DeriveLink ...
func deriveLink(target, targetDir, linkDir string) (string, error) {
	relToDotfiles, err := filepath.Rel(targetDir, target)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"target":    target,
			"targetDir": targetDir,
		}).WithError(err).Error("unable to make target relative to target dir")
		return "", err
	}

	if strings.HasPrefix(relToDotfiles, "..") {
		return "", ErrNonHomeRelativeTarget
	}

	return filepath.Join(linkDir, relToDotfiles), nil
}

// Ensure the existence of the symlink. If the symlink already
// exists this does nothing. Otherwise it will create a symlink from
// link to target.
//
// If the given symlink has an existing file at link but not target this
// will be treated as a file to add, meaning the file at link will be moved
// to the target path before creating the symlink from link to target.
func (symlink *Symlink) Ensure(config conf.Config) error {
	logger := logrus.WithFields(logrus.Fields{
		"link":   symlink.Link,
		"target": symlink.Target,
	})

	if symlink.Exists(config) {
		return nil
	}

	linkExists := false
	if _, err := config.Fs.Stat(symlink.Link); err == nil {
		linkExists = true
	}

	if linkExists {
		if symlink.Target == "" {
			logger.Debug("no target given, deriving from link")
			target, err := deriveLink(symlink.Link, config.UserHome, config.PunktHome)
			if err != nil {
				return err
			}

			symlink.Target = target
			if symlink.Exists(config) {
				return nil
			}
		}

		targetExists := false
		if _, err := config.Fs.Stat(symlink.Target); err == nil {
			targetExists = true
		}

		if !targetExists {
			err := path.CreateNecessaryDirectories(config.Fs, symlink.Target)
			if err != nil {
				return err
			}

			logger.Debug("target doesn't exist, assuming link is the target")
			err = config.Fs.Rename(symlink.Link, symlink.Target)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"symlink": symlink,
				}).WithError(err).Error("unable to move link to target location")
				return err
			}
		}
	}

	err := path.CreateNecessaryDirectories(config.Fs, symlink.Link)
	if err != nil {
		return err
	}

	logger.Info("creating symlink")
	return config.Fs.Symlink(symlink.Target, symlink.Link)
}

// Exists returns true if there exists a symlink at Link pointing to Target
func (symlink Symlink) Exists(config conf.Config) bool {
	logger := logrus.WithFields(logrus.Fields{
		"target": symlink.Target,
		"link":   symlink.Link,
	})
	logger.Debug("checking if symlink exists")

	if _, err := config.Fs.Lstat(symlink.Link); err != nil {
		logger.WithError(err).Debug("failed to Lstat link")
		return false
	}

	path, err := config.Fs.Readlink(symlink.Link)
	if err != nil {
		logger.WithError(err).Debug("unable to readlink")
		return false
	}

	return path == symlink.Target
}
