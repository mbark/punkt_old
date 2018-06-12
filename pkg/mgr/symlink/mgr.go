package symlink

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/mbark/punkt/pkg/conf"
	"github.com/mbark/punkt/pkg/fs"
	"github.com/mbark/punkt/pkg/printer"
)

// Manager ...
type Manager struct {
	LinkManager LinkManager
	snapshot    fs.Snapshot
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

func (symlink Symlink) String() string {
	return fmt.Sprintf("%s -> %s", symlink.Link, symlink.Target)
}

// UnmarshalTOML unmarshals a map of link -> target
func (config *Config) UnmarshalTOML(data interface{}) error {
	links, _ := data.(map[string]interface{})
	for link, val := range links {
		target, _ := val.(string)

		s := Symlink{
			Link:   link,
			Target: target,
		}

		config.Symlinks = append(config.Symlinks, s)
	}

	return nil
}

// AsMap returns the configuration as a map, which is the format the
// symlinks should be stored in..
func (config Config) AsMap() map[string]string {
	mapping := make(map[string]string)
	for _, s := range config.Symlinks {
		mapping[s.Link] = s.Target
	}
	return mapping
}

// NewManager ...
func NewManager(c conf.Config, snapshot fs.Snapshot, configFile string) *Manager {
	return &Manager{
		LinkManager: NewLinkManager(c, snapshot),
		snapshot:    snapshot,
		config:      c,
		configFile:  configFile,
	}
}

// Add ...
func (mgr Manager) Add(target, newLocation string) (*Symlink, error) {
	absTarget, err := mgr.snapshot.AsAbsolute(target)
	if err != nil {
		printer.Log.Error("target file or directory does not exist: {fg 1}%s", target)
		return nil, err
	}

	symlink := mgr.LinkManager.New(newLocation, absTarget)
	err = mgr.LinkManager.Ensure(symlink)
	if err != nil {
		printer.Log.Error("failed to create symlink: {fg 1}%s", err)
		return nil, errors.Wrapf(err, "failed to ensure %s exists", symlink)
	}

	storedLink, err := mgr.addToConfiguration(symlink)
	if err == nil {
		printer.Log.Success("symlink added: {fg 2}%s", storedLink)
	} else {
		printer.Log.Error("failed to add symlink: {fg 1}%s", err)
	}

	return symlink, err
}

// Remove ...
func (mgr Manager) Remove(link string) error {
	absLink, err := mgr.snapshot.AsAbsolute(link)
	if err != nil {
		printer.Log.Error("file does not exist: {fg 1}%s", link)
		return err
	}

	s, err := mgr.LinkManager.Remove(absLink)
	if err != nil {
		printer.Log.Error("failed to remove link, error was: {fg 1}%s", err)
		err = errors.Wrapf(err, "failed to remove link %s", link)
		return err
	}

	removedLink, err := mgr.removeFromConfiguration(*s)
	if err == nil {
		if removedLink != nil {
			printer.Log.Success("symlink removed: {fg 2}%s", removedLink)
		}
	} else {
		printer.Log.Error("failed to remove symlink: {fg 1}%s", err)
	}

	return err
}

func (mgr Manager) readConfiguration() (Config, error) {
	var savedConfig Config
	err := mgr.snapshot.ReadToml(&savedConfig, mgr.configFile)
	if err != nil {
		logger := logrus.WithField("configFile", mgr.configFile).WithError(err)
		if err == fs.ErrNoSuchFile {
			printer.Log.Note("no symlink configuration file at {fg 5}%s", mgr.snapshot.UnexpandHome(mgr.configFile))
			logger.Warn("no configuration file found")
		} else {
			logger.Error("unable to read symlink configuration file")
		}
	}

	return savedConfig, err
}

func (mgr Manager) addToConfiguration(new *Symlink) (*Symlink, error) {
	logrus.WithField("newSymlink", new).Info("Storing symlink in configuration")
	saved, err := mgr.readConfiguration()
	if err != nil && err != fs.ErrNoSuchFile {
		return nil, err
	}

	unexpanded := mgr.LinkManager.Unexpand(*new)
	for _, existing := range saved.Symlinks {
		if unexpanded.Target == existing.Target && unexpanded.Link == existing.Link {
			printer.Log.Note("symlink is already stored")
			logrus.WithField("symlink", unexpanded).Info("symlink already saved, nothing new to store")
			return unexpanded, nil
		}
	}

	saved.Symlinks = append(saved.Symlinks, *unexpanded)

	logrus.WithField("symlinks", saved).Debug("storing updated list of symlinks")
	return unexpanded, mgr.snapshot.SaveToml(saved.AsMap(), mgr.configFile)
}

func (mgr Manager) removeFromConfiguration(symlink Symlink) (*Symlink, error) {
	var config Config
	err := mgr.snapshot.ReadToml(&config, mgr.configFile)
	if err == fs.ErrNoSuchFile {
		logrus.WithFields(logrus.Fields{
			"configFile": mgr.configFile,
		}).WithError(err).Warn("no configuration file found, configuration won't be updated")
		printer.Log.Warning("no symlink configuration found")
		return nil, nil
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
		printer.Log.Warning("symlink not found in configuration, nothing to remove")
		return unexpanded, nil
	}

	config.Symlinks = append(config.Symlinks[:index], config.Symlinks[index+1:]...)
	return unexpanded, mgr.snapshot.SaveToml(config.AsMap(), mgr.configFile)
}

// Name ...
func (mgr Manager) Name() string {
	return "symlink"
}

// Dump ...
func (mgr Manager) Dump() (string, error) { return "", nil }

// Update ...
func (mgr Manager) Update() error { return nil }

// Ensure ...
func (mgr Manager) Ensure() error { return nil }
