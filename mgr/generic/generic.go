package generic

import (
	"os/exec"
	"strings"

	"github.com/mbark/punkt/conf"
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
func NewManager(c conf.Config, configFile, name string) *Manager {
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
	logger := logrus.WithFields(logrus.Fields{
		"operation": operation,
		"args":      args,
	})

	if val, ok := mgr.commands[operation]; ok {
		logger.Info("operation found in manager config")
		name = val
	} else {
		logger.WithField("command", mgr.commands).Info("operation not found in manager config, using 'command'")
		name = mgr.commands["command"]
		args = append([]string{operation}, args...)
	}

	command := strings.Join(append([]string{name}, args...), " ")
	logger.WithField("command", command).Info("resolved command to use")

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
