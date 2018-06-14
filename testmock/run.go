package testmock

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/mbark/punkt/pkg/run"
)

const wantHelper = "GO_WANT_HELPER_PROCESS"

// NoOutput can be used to make sure that the command's output is discarded
func NoOutput() {
	run.Out = ioutil.Discard
}

// FakeWithEnvCommand creates an exec.Cmd that can be run with an environment
// variable set. This is a useful way to pass information to the process on
// how it should behave.
func FakeWithEnvCommand(helper, env string) func(string, ...string) *exec.Cmd {
	return func(command string, args ...string) *exec.Cmd {
		cmd := FakeCommand(helper)(command, args...)
		cmd.Env = append(cmd.Env, env)
		return cmd
	}
}

// FakeCommand creates an exec.Cmd that runs the given helper method.
func FakeCommand(helper string) func(string, ...string) *exec.Cmd {
	return func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=" + helper, "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{wantHelper + "=1"}
		return cmd
	}
}

// VerifyHelperProcess can be used at the start of your helper process to
// make sure that the process is run via FakeCommand and not as a test.
// If the process is run as a test an error will be returned, otherwise
// the command and arguments will be given.
func VerifyHelperProcess() (string, []string, error) {
	if os.Getenv(wantHelper) != "1" {
		return "", []string{}, errors.New("ran as a test, ignoring")
	}

	if os.Getenv("FAILING") == "true" {
		os.Exit(3)
	}

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}

		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "no command\n")
		os.Exit(2)
	}

	return args[0], args[1:], nil
}
