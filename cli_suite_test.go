package main_test

import (
	"bytes"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPunkt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Suite")
}

var _ = Describe("CLI-API", func() {
	var _ = BeforeEach(func() {
		cmd := exec.Command("go", "build")
		expectSuccess(cmd)
	})

	It("should have --help for all commands", func() {
		for _, command := range []string{"", "add", "ensure", "dump", "update"} {
			cmd := exec.Command("./punkt", command, "--help")
			expectSuccess(cmd)
		}
	})

	It("should have a --version command", func() {
		cmd := exec.Command("./punkt", "--version")
		expectSuccess(cmd)
	})
})

func expectSuccess(cmd *exec.Cmd) {
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	Expect(cmd.Run()).To(BeNil(), stderr.String())
}
