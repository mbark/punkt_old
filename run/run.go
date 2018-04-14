package run

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

// PrintOutputToUser modifies the given command so that std{out,err,in} will use
// the system-wide ones, thus printing it directly to the terminal. This should
// be used when you want to show what is happening to the user.
func PrintOutputToUser(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
}

// CaptureOutput captures std{out,err} of the command in the byte buffers
// returned. This is useful when you want to use the output of the command.
func CaptureOutput(cmd *exec.Cmd) (*bytes.Buffer, *bytes.Buffer) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	return &stdout, &stderr
}

func loggerWithCmd(cmd *exec.Cmd) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"cmd": strings.Join(cmd.Args, " "),
	})
}

// WithCapture runs the given command while printing the output
// to the user and also capturing stdout for later use.
func WithCapture(cmd *exec.Cmd) (*bytes.Buffer, error) {
	logger := loggerWithCmd(cmd)
	var stdoutBuf bytes.Buffer

	stdoutIn, _ := cmd.StdoutPipe()
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	stdout := io.MultiWriter(os.Stdout, &stdoutBuf)

	err := cmd.Start()
	if err != nil {
		logger.WithError(err).Error("Failed to start command")
		return nil, err
	}

	var stdoutErr error
	go func() {
		_, stdoutErr = io.Copy(stdout, stdoutIn)
		if stdoutErr != nil {
			logrus.WithError(err).Error("stdout copy failed")
		}
	}()

	err = cmd.Wait()
	if err != nil {
		logger.WithError(err).Error("Unable to run command")
	}

	return &stdoutBuf, nil
}

// Run ...
func Run(cmd *exec.Cmd) error {
	logger := loggerWithCmd(cmd)

	logger.Info("Running command")
	err := cmd.Run()

	if err != nil {
		logger.WithError(err).Error("Unable to run command")
		return err
	}

	logger.WithField("rawCmd", cmd).Debug("Command finished without error")
	return nil
}
