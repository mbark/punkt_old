package symlink

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-billy.v4"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/file"
	"github.com/mbark/punkt/path"
)

// Manager ...
type Manager struct {
	config conf.Config
}

// NewManager ...
func NewManager(c conf.Config) *Manager {
	return &Manager{
		config: c,
	}
}

// Symlink describes a symlink, i.e. what it links from and what it links to
type Symlink struct {
	fs   billy.Filesystem
	From string
	To   string
}

// NewSymlink creates a new symlink
func NewSymlink(fs billy.Filesystem, from, to string) *Symlink {
	return &Symlink{
		fs:   fs,
		From: from,
		To:   to,
	}
}

func (symlink Symlink) expand() Symlink {
	return *NewSymlink(symlink.fs, path.ExpandHome(symlink.From), path.ExpandHome(symlink.To))
}

func (symlink Symlink) unexpend() Symlink {
	return *NewSymlink(symlink.fs, path.UnexpandHome(symlink.From), path.UnexpandHome(symlink.To))
}

// Create will construct the corresponding symlink. Returns true if the symlink
// was successfully created, otherwise false.
func (symlink Symlink) Create() error {
	logger := logrus.WithFields(logrus.Fields{
		"to":   symlink.To,
		"from": symlink.From,
	})

	_, err := symlink.fs.Stat(symlink.From)
	if err != nil {
		return err
	}

	err = path.CreateNecessaryDirectories(symlink.fs, symlink.To)
	if err != nil {
		logger.WithError(err).Error("Unable to create necessary directories")
		return err
	}

	logger.Info("Creating symlink")
	return symlink.fs.Symlink(symlink.From, symlink.To)
}

// Exists returns true if the symlink already exists
func (symlink Symlink) Exists() bool {
	from, _ := symlink.fs.Stat(symlink.From)
	to, _ := symlink.fs.Stat(symlink.To)

	logrus.WithFields(logrus.Fields{
		"to":   symlink.From,
		"from": symlink.To,
	}).Debug("Comparing if files are the same")
	return os.SameFile(from, to)
}

// Dump ...
func (mgr Manager) Dump() error { return nil }

// Update ...
func (mgr Manager) Update() error { return nil }

// Ensure goes through the list of symlinks ensuring they exist
func (mgr Manager) Ensure() error {
	symlinks := []Symlink{}
	file.Read(mgr.config.Fs, &symlinks, mgr.config.Dotfiles, "symlinks")

	failed := []Symlink{}
	for _, symlink := range symlinks {
		s := symlink.expand()

		if s.Exists() {
			logrus.WithFields(logrus.Fields{
				"from": s.From,
				"to":   s.To,
			}).Debug("Symlink already exists, not creating")
		} else {
			err := s.Create()
			if err != nil {
				logrus.WithField("symlink", symlink).WithError(err).Error("Failed to create symlink")
				failed = append(failed, symlink)
			}
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("The following symlinks could not be created: %v", failed)
	}

	return nil
}
