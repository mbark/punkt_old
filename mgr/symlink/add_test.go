package symlink_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/file"
	"github.com/mbark/punkt/mgr/symlink"
)

func TestAdd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Symlink Suite")
}

var _ = Describe("Symlink: Add", func() {
	var fs billy.Filesystem
	var config conf.Config
	var mgr *symlink.Manager
	var testfile string

	BeforeEach(func() {
		fs = memfs.New()

		userhome := "/home"
		punktHome := userhome + "/.config/punkt"
		dotfiles := userhome + "/dotfiles"

		err := fs.MkdirAll(punktHome, os.ModePerm)
		ExpectNoErr(err)

		err = fs.MkdirAll(dotfiles, os.ModePerm)
		ExpectNoErr(err)

		testfile = userhome + "/testfile"
		f, err := fs.Create(testfile)
		ExpectNoErr(err)

		_, err = f.Write([]byte("foo"))
		ExpectNoErr(err)

		err = f.Close()
		ExpectNoErr(err)

		_, err = fs.Stat(testfile)
		ExpectNoErr(err)

		config = conf.Config{
			PunktHome:  punktHome,
			Dotfiles:   dotfiles,
			UserHome:   userhome,
			Fs:         fs,
			WorkingDir: userhome,
		}

		mgr = symlink.NewManager(config)
		logrus.SetLevel(logrus.PanicLevel)
	})

	It("should give an error for a non-existant file", func() {
		Expect(mgr.Add("nonExistantFile", "")).NotTo(Succeed())
	})

	It("should be able to add a symlink a file", func() {
		Expect(mgr.Add(testfile, "")).To(Succeed())

		info, err := fs.Lstat(testfile)
		ExpectNoErr(err)
		Expect(info.Mode() & os.ModeSymlink).NotTo(Equal(0))
	})

	It("should create a symlinks.yaml file when adding a symlink", func() {
		Expect(mgr.Add(testfile, "")).To(Succeed())

		_, err := fs.Stat(config.Dotfiles + "/symlinks.yml")
		ExpectNoErr(err)
	})

	It("should add the symlink to the symlinks.yml file", func() {
		Expect(mgr.Add(testfile, "")).To(Succeed())

		actual := []symlink.Symlink{}
		err := file.Read(fs, &actual, config.Dotfiles, "symlinks")
		ExpectNoErr(err)

		expected := []symlink.Symlink{
			{
				From: config.Dotfiles + "/testfile",
				To:   testfile,
			},
		}

		Expect(actual).Should(Equal(expected))
	})
})

func ExpectNoErr(err error) {
	if err != nil {
		Expect(err).To(BeNil())
	}
}
