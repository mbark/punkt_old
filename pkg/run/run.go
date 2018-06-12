package run

import (
	"io"
	"os"
	"os/exec"
)

// Commander is the function used to create the command to run, you can
// set this variable to some other way to construct arguments if you
// want to mock how commands are run.
var Commander = exec.Command

// Out is the output to use when printing to the user. By default this is
// os.Stdout but can be changed, e.g. when running tests that you don't
// want printing output.
var Out io.Writer = os.Stdout

// PrintToUser sets up the command to print all output the user and to use
// os.Stdin to capture input.
func PrintToUser(cmd *exec.Cmd) {
	cmd.Stdout = Out
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
}
