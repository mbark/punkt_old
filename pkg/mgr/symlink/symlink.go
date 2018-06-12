package symlink

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/mbark/punkt/pkg/conf"
	"github.com/mbark/punkt/pkg/fs"
	"github.com/mbark/punkt/pkg/printer"
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
	snapshot fs.Snapshot
	config   conf.Config
}

// NewLinkManager ...
func NewLinkManager(config conf.Config, snapshot fs.Snapshot) LinkManager {
	return symlinkManager{
		snapshot: snapshot,
		config:   config,
	}
}

// New ...
func (mgr symlinkManager) New(target, link string) *Symlink {
	logger := logrus.WithFields(logrus.Fields{
		"target": target,
		"link":   link,
	})
	logger.Debug("new symlink")

	home := mgr.snapshot.UserHome
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
	target, err := mgr.snapshot.Fs.Readlink(link)
	if err != nil {
		return nil, errors.Wrapf(err, "given link isn't a symlink")
	}

	err = mgr.snapshot.Fs.Remove(link)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to remove %s", link)
	}

	err = mgr.snapshot.Fs.Rename(target, link)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to move %s to %s location", target, link)
	}

	return mgr.New(target, link), nil
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
		printer.Log.Note("symlink exists: {fg 5}%s", mgr.Unexpand(*symlink))
		return nil
	}

	linkexists := false
	if _, err := mgr.snapshot.Fs.Stat(symlink.Link); err == nil {
		linkexists = true
	}

	targetexists := false
	if _, err := mgr.snapshot.Fs.Stat(symlink.Target); err == nil {
		targetexists = true
	}

	logger.WithFields(logrus.Fields{
		"linkExists":   linkexists,
		"targetExists": targetexists,
	}).Debug("status of symlink")

	if linkexists && !targetexists {
		logger.Debug("link exists but target doesn't, moving link -> target")

		err := mgr.snapshot.CreateNecessaryDirectories(symlink.Target)
		if err != nil {
			return err
		}

		err = mgr.snapshot.Fs.Rename(symlink.Link, symlink.Target)
		if err != nil {
			logger.WithError(err).Error("unable to move link to target location")
			return errors.Wrapf(err, "failed to rename %s to %s", symlink.Link, symlink.Target)
		}
	}

	err := mgr.snapshot.CreateNecessaryDirectories(symlink.Link)
	if err != nil {
		return err
	}

	logger.Info("creating symlink")
	printer.Log.Note("creating symlink: {fg 2}%s", mgr.Unexpand(*symlink))
	return mgr.snapshot.Fs.Symlink(symlink.Target, symlink.Link)
}

// exists returns true if there exists a symlink at Link pointing to Target
func (mgr symlinkManager) exists(symlink *Symlink) bool {
	logger := logrus.WithFields(logrus.Fields{
		"target": symlink.Target,
		"link":   symlink.Link,
	})
	logger.Debug("checking if symlink exists")

	path, err := mgr.snapshot.Fs.Readlink(symlink.Link)
	if err != nil {
		logger.WithError(err).Debug("unable to readlink")
		return false
	}

	return path == symlink.Target
}

// Expand ...
func (mgr symlinkManager) Expand(symlink Symlink) *Symlink {
	return &Symlink{
		Target: mgr.snapshot.ExpandHome(symlink.Target),
		Link:   mgr.snapshot.ExpandHome(symlink.Link),
	}
}

// Unexpand ...
func (mgr symlinkManager) Unexpand(symlink Symlink) *Symlink {
	return &Symlink{
		Target: mgr.snapshot.UnexpandHome(symlink.Target),
		Link:   mgr.snapshot.UnexpandHome(symlink.Link),
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
