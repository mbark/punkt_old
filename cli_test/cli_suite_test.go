package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	// "os"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var dotfiles string

func TestPunkt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Suite")
}

var _ = BeforeSuite(func() {
	// Ensure punkt is built
	cmd := exec.Command("go", "build")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Dir = ".."
	Expect(cmd.Run()).To(BeNil(), stderr.String())

	// Create a temporary dotfiles directory
	name, err := ioutil.TempDir("", "dotfiles")
	Expect(err).To(BeNil())
	dotfiles = name
})

var _ = AfterSuite(func() {
	// os.RemoveAll(dotfiles)
})

type Punkt struct {
	cmd    *exec.Cmd
	stdout *bytes.Buffer
	stderr *bytes.Buffer
}

func NewPunkt(subcommand string, args ...string) *Punkt {
	arguments := append([]string{subcommand, "--dotfiles", dotfiles}, args...)
	cmd := exec.Command("../punkt", arguments...)
	var stderr bytes.Buffer
	var stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	return &Punkt{
		cmd:    cmd,
		stdout: &stdout,
		stderr: &stderr,
	}
}

func (punkt *Punkt) ExpectSuccess() {
	description := fmt.Sprintf("STDOUT:\n%s\nSTDERR:\n%s\nCOMMAND:\n%v\n", punkt.stdout.String(), punkt.stderr.String(), punkt.cmd.Args)
	Expect(punkt.cmd.Run()).To(BeNil(), description)
}
