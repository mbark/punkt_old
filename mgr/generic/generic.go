package generic

import (
	"os/exec"
	"strings"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/file"
	"github.com/mbark/punkt/mgr/symlink"
	"github.com/mbark/punkt/run"
	"github.com/sirupsen/logrus"
)

// Manager ...
type Manager struct {
	name       string
	config     conf.Config
	commands   map[string]string
	configFile string
}

// Config ...
type Config struct {
	Symlinks []symlink.Symlink
}

// NewManager ...
func NewManager(name, configFile string, c conf.Config) *Manager {
	logrus.WithFields(logrus.Fields{
		"name":     name,
		"commands": c.Managers[name],
	}).Info("Constructing generic manager")

	return &Manager{
		name:       name,
		config:     c,
		commands:   c.Managers[name],
		configFile: configFile,
	}
}

func (mgr Manager) resolveCommand(operation string, args ...string) *exec.Cmd {
	var name string

	if val, ok := mgr.commands[operation]; ok {
		name = val
	} else {
		name = mgr.commands["command"]
		args = append([]string{operation}, args...)
	}

	logrus.WithFields(logrus.Fields{
		"operation": operation,
		"args":      args,
		"command":   name,
	}).Info("resolved command to use")

	command := strings.Join(append([]string{name}, args...), " ")
	return mgr.config.Command("sh", "-c", command)
}

// Name ...
func (mgr Manager) Name() string {
	return mgr.name
}

// Dump ...
func (mgr Manager) Dump() (string, error) {
	cmd := mgr.resolveCommand("dump")
	stdout, err := run.WithCapture(cmd)
	if err != nil {
		return "", err
	}

	return stdout.String(), nil
}

// Update ...
func (mgr Manager) Update() error {
	cmd := mgr.resolveCommand("ensure", mgr.configFile)
	run.PrintOutputToUser(cmd)

	return run.Run(cmd)
}

// Ensure ...
func (mgr Manager) Ensure() error {
	cmd := mgr.resolveCommand("ensure", mgr.configFile)
	run.PrintOutputToUser(cmd)

	return run.Run(cmd)
}

// Symlinks ...
func (mgr Manager) Symlinks() ([]symlink.Symlink, error) {
	var config Config
	err := file.ReadToml(mgr.config.Fs, &config, mgr.configFile)
	if err != nil && err != file.ErrNoSuchFile {
		if err == file.ErrNoSuchFile {
			return []symlink.Symlink{}, nil
		}

		return nil, err
	}

	return config.Symlinks, nil
}
