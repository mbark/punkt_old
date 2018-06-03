package symlink

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/path"
)

var (
	// ErrNonHomeRelativeTarget is returned if a target has no given link and is outside of the user's home directory
	ErrNonHomeRelativeTarget = errors.New("non-home relative target given without specific link location")
)

// LinkManager ...
type LinkManager interface {
	New(target, link string) *Symlink
	Remove(string) (*Symlink, error)
	Ensure(symlink *Symlink) error
	Unexpand(symlink Symlink) *Symlink
	Expand(symlink Symlink) *Symlink
}

type symlinkManager struct {
	config conf.Config
}

// NewLinkManager ...
func NewLinkManager(config conf.Config) LinkManager {
	return symlinkManager{
		config: config,
	}
}

// New ...
func (mgr symlinkManager) New(target, link string) *Symlink {
	logger := logrus.WithFields(logrus.Fields{
		"target": target,
		"link":   link,
	})
	logger.Debug("new symlink")

	home := mgr.config.UserHome
	dotfiles := mgr.config.Dotfiles

	if target == "" && link != "" {
		logger.Debug("empty target with non-empty link, deriving target")
		target, _ = deriveLink(link, home, dotfiles)
	}

	if target != "" && link == "" {
		logger.Debug("empty link with non-empty target, deriving link")
		link, _ = deriveLink(target, dotfiles, home)
	}

	return &Symlink{
		Target: target,
		Link:   link,
	}
}

// Remove ...
func (mgr symlinkManager) Remove(link string) (*Symlink, error) {
	target, err := mgr.config.Fs.Readlink(link)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read link [link: %s]", link)
	}

	symlink := mgr.New(target, link)

	err = mgr.config.Fs.Remove(link)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to remove file [link: %s]", link)
	}

	return symlink, mgr.config.Fs.Rename(target, link)
}

// Ensure the existence of the symlink. If the symlink already
// exists this does nothing. Otherwise it will create a symlink from
// link to target.
//
// If the given symlink has an existing file at link but not target this
// will be treated as a file to add, meaning the file at link will be moved
// to the target path before creating the symlink from link to target.
func (mgr symlinkManager) Ensure(symlink *Symlink) error {
	logger := logrus.WithFields(logrus.Fields{
		"link":   symlink.Link,
		"target": symlink.Target,
	})

	if mgr.exists(symlink) {
		return nil
	}

	linkexists := false
	if _, err := mgr.config.Fs.Stat(symlink.Link); err == nil {
		linkexists = true
	}

	targetexists := false
	if _, err := mgr.config.Fs.Stat(symlink.Target); err == nil {
		targetexists = true
	}

	if !linkexists && !targetexists {
		return errors.Errorf("neither link nor target exists [link: %s, target: %s]", symlink.Link, symlink.Target)
	}

	if linkexists && !targetexists {
		err := path.CreateNecessaryDirectories(mgr.config.Fs, symlink.Target)
		if err != nil {
			return errors.Wrapf(err, "failed to create necessary directories [path: %s]", symlink.Target)
		}

		logger.Debug("target doesn't exist, assuming link is the target")
		err = mgr.config.Fs.Rename(symlink.Link, symlink.Target)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"symlink": symlink,
			}).WithError(err).Error("unable to move link to target location")
			return errors.Wrapf(err, "failed to rename link to target [link: %s, target: %s]", symlink.Link, symlink.Target)
		}
	}

	err := path.CreateNecessaryDirectories(mgr.config.Fs, symlink.Link)
	if err != nil {
		return errors.Wrapf(err, "failed to create necessary directories [path: %s]", symlink.Link)
	}

	logger.Info("creating symlink")
	return mgr.config.Fs.Symlink(symlink.Target, symlink.Link)
}

// exists returns true if there exists a symlink at Link pointing to Target
func (mgr symlinkManager) exists(symlink *Symlink) bool {
	logger := logrus.WithFields(logrus.Fields{
		"target": symlink.Target,
		"link":   symlink.Link,
	})
	logger.Debug("checking if symlink exists")

	path, err := mgr.config.Fs.Readlink(symlink.Link)
	if err != nil {
		logger.WithError(err).Debug("unable to readlink")
		return false
	}

	return path == symlink.Target
}

// Expand ...
func (mgr symlinkManager) Expand(symlink Symlink) *Symlink {
	home := mgr.config.UserHome
	return &Symlink{
		Target: path.ExpandHome(symlink.Target, home),
		Link:   path.ExpandHome(symlink.Link, home),
	}
}

// Unexpand ...
func (mgr symlinkManager) Unexpand(symlink Symlink) *Symlink {
	home := mgr.config.UserHome

	return &Symlink{
		Target: path.UnexpandHome(symlink.Target, home),
		Link:   path.UnexpandHome(symlink.Link, home),
	}
}

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
