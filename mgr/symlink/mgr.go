package symlink

import (
	"fmt"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/file"
)

// Manager ...
type Manager struct {
	LinkManager LinkManager
	configFile  string
	config      conf.Config
}

// Symlink describes a symlink, i.e. what it links from and what it links to
type Symlink struct {
	Target string
	Link   string
}

// Config ...
type Config struct {
	Symlinks []Symlink
}

// NewManager ...
func NewManager(c conf.Config, configFile string) *Manager {
	return &Manager{
		LinkManager: NewLinkManager(c),
		config:      c,
		configFile:  configFile,
	}
}

// Add ...
func (mgr Manager) Add(target, newLocation string) (*Symlink, error) {
	if !filepath.IsAbs(target) {
		target = mgr.config.Fs.Join(mgr.config.WorkingDir, target)
	}

	if newLocation == "" {
		loc, err := deriveLink(target, mgr.config.UserHome, mgr.config.Dotfiles)
		if err != nil {
			return nil, err
		}

		newLocation = loc
	}

	symlink := mgr.LinkManager.New(newLocation, target)
	err := mgr.LinkManager.Ensure(symlink)
	if err != nil {
		return symlink, err
	}

	return symlink, mgr.saveSymlink(symlink)
}

func (mgr Manager) saveSymlink(new *Symlink) error {
	var saved Config
	err := file.ReadToml(mgr.config.Fs, &saved, mgr.configFile)
	if err != nil && err != file.ErrNoSuchFile {
		logrus.WithError(err).WithField("symlink", new).Error("unable to read file containing all symlinks, assuming non exists")
		return err
	}

	unexpanded := mgr.LinkManager.Unexpand(*new)
	for _, existing := range saved.Symlinks {
		if unexpanded.Target == existing.Target && unexpanded.Link == existing.Link {
			logrus.WithField("symlink", unexpanded).Info("symlink already saved, nothing new to store")
			return nil
		}
	}

	saved.Symlinks = append(saved.Symlinks, *unexpanded)
	logrus.WithFields(logrus.Fields{
		"symlinks": saved,
	}).Debug("Storing updated list of symlinks")
	return file.SaveToml(mgr.config.Fs, saved, mgr.configFile)
}

// Remove ...
func (mgr Manager) Remove(link string) error {
	s, err := mgr.LinkManager.Remove(link)
	if err != nil {
		return err
	}

	return mgr.removeFromConfiguration(*s)
}

func (mgr Manager) removeFromConfiguration(symlink Symlink) error {
	var config Config
	err := file.ReadToml(mgr.config.Fs, &config, mgr.configFile)
	if err == file.ErrNoSuchFile {
		logrus.WithFields(logrus.Fields{
			"configFile": mgr.configFile,
		}).WithError(err).Warn("no configuration file found, configuration won't be updated")
		return nil
	}

	unexpanded := mgr.LinkManager.Unexpand(symlink)
	index := -1
	for i, s := range config.Symlinks {
		logrus.WithFields(logrus.Fields{
			"unexpanded": unexpanded,
			"saved":      s,
		}).Debug("comparing if symlinks are the same")
		if unexpanded.Target == s.Target && unexpanded.Link == s.Link {
			index = i
		}
	}

	if index < 0 {
		logrus.WithFields(logrus.Fields{
			"symlink": symlink,
			"config":  config,
		}).Warn("symlink not found in configuration, not removing")
		return nil
	}

	config.Symlinks = append(config.Symlinks[:index], config.Symlinks[index+1:]...)
	return file.SaveToml(mgr.config.Fs, config, mgr.configFile)
}

// Name ...
func (mgr Manager) Name() string {
	return "symlink"
}

// Dump ...
func (mgr Manager) Dump() (string, error) { return "", nil }

// Update ...
func (mgr Manager) Update() error { return nil }

// Ensure goes through the list of symlinks ensuring they exist
func (mgr Manager) Ensure() error {
	var config Config
	err := file.ReadToml(mgr.config.Fs, &config, mgr.configFile)
	if err != nil {
		if err == file.ErrNoSuchFile {
			return nil
		}

		return err
	}

	failed := []Symlink{}
	for _, symlink := range config.Symlinks {
		s := mgr.LinkManager.New(symlink.Target, symlink.Link)
		err = mgr.LinkManager.Ensure(s)
		if err != nil {
			logrus.WithField("symlink", symlink).WithError(err).Error("failed to ensure symlink")
			failed = append(failed, symlink)
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("The following symlinks could not be created: %v", failed)
	}

	return nil
}

// Symlinks ...
func (mgr Manager) Symlinks() ([]Symlink, error) {
	var config Config
	err := file.ReadToml(mgr.config.Fs, &config, mgr.configFile)
	if err != nil && err != file.ErrNoSuchFile {
		if err == file.ErrNoSuchFile {
			return []Symlink{}, nil
		}

		return nil, err
	}

	return config.Symlinks, nil
}
