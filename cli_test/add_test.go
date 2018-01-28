package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	"github.com/mbark/punkt/mgr/symlink"
	"github.com/mbark/punkt/path"
)

var _ = Describe("Add", func() {
	var from *os.File
	var symlinks string

	BeforeEach(func() {
		var err error
		from, err = os.Create("fromtarget")
		Expect(err).To(BeNil())
		Expect(from).NotTo(BeNil())

		symlinks = filepath.Join(dotfiles, "symlinks.yml")
	})

	AfterEach(func() {
		os.Remove(from.Name())
		os.Remove(symlinks)
	})

	It("should give an error for a non-existant file", func() {
		punkt := NewPunkt("add", "nonExistantFile")
		Expect(punkt.cmd.Run()).NotTo(BeNil(), punkt.stderr.String())
	})

	It("should be able to add a symlink a file", func() {
		punkt := NewPunkt("add", from.Name())
		punkt.ExpectSuccess()

		from, err := os.Stat(from.Name())
		Expect(err).To(BeNil())
		Expect(from.Mode() & os.ModeSymlink).NotTo(Equal(0))
	})

	It("should create a symlinks.yaml file when adding a symlink", func() {
		punkt := NewPunkt("add", from.Name())
		punkt.ExpectSuccess()

		_, err := os.Stat(symlinks)
		if err != nil {
			Expect(err).To(BeNil(), err.Error())
		}
	})

	It("should add the symlink to the symlinks.yml file", func() {
		punkt := NewPunkt("add", from.Name())
		punkt.ExpectSuccess()

		content, err := ioutil.ReadFile(symlinks)
		Expect(err).To(BeNil())

		f, err := filepath.Abs(from.Name())
		Expect(err).To(BeNil())

		rel, err := filepath.Rel(path.GetUserHome(), f)
		Expect(err).To(BeNil())
		to := filepath.Join(dotfiles, rel)

		expected, err := yaml.Marshal([]symlink.Symlink{
			{
				From: path.UnexpandHome(f),
				To:   to,
			},
		})
		Expect(err).To(BeNil())

		Expect(content).Should(MatchYAML(expected))
	})
})
