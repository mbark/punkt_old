package cmd

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("CLI-API", func() {
	It("should have --help for all commands", func() {
		for _, command := range []string{"", "add", "ensure", "dump", "update"} {
			punkt := NewPunkt(command, "--help")
			punkt.ExpectSuccess()
		}
	})
})
