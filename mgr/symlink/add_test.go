package symlink_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/file"
	"github.com/mbark/punkt/mgr/symlink"
	"github.com/mbark/punkt/path"
)

var _ = Describe("Symlink: Add", func() {
	var fs billy.Filesystem
	var config conf.Config
	var mgr *symlink.Manager
	var testfile string
	var configFile string

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
		configFile = filepath.Join(config.PunktHome, "symlinks.toml")

		mgr = symlink.NewManager(config, configFile)
		logrus.SetLevel(logrus.PanicLevel)
	})

	It("should be possible to create a symlink", func() {
		_, err := mgr.Add(testfile, "")
		Expect(err).To(Succeed())

		info, err := fs.Lstat(testfile)
		ExpectNoErr(err)
		Expect(info.Mode() & os.ModeSymlink).NotTo(Equal(0))

		f, err := fs.Readlink(testfile)
		ExpectNoErr(err)
		Expect(f).To(Equal(config.Dotfiles + "/testfile"))
	})

	It("should fail to add a file if the new location for it already exists", func() {
		_, err := fs.Create(config.Dotfiles + "/testfile")
		ExpectNoErr(err)

		_, err = mgr.Add(testfile, "")
		Expect(err).NotTo(Succeed())
	})

	Context("when adding symlinks", func() {
		var link string

		BeforeEach(func() {
			link = filepath.Join(config.UserHome, "link")
			err := fs.Symlink(testfile, link)
			Expect(err).To(BeNil())
		})

		It("should not exist initially", func() {
			s, err := mgr.Add(link, "")
			Expect(err).To(Succeed())

			Expect(s.Exists(config)).To(BeFalse())
		})
	})

	Context("when reading and saving to the symlinks.yml file", func() {
		var initial symlink.Config
		var testfileSymlink symlink.Symlink

		var expectNewSymlink func(symlink.Symlink)

		BeforeEach(func() {
			initial = symlink.Config{
				Symlinks: []symlink.Symlink{
					{
						Target: "~/some/file",
						Link:   "~/to/some/file",
					},
				},
			}
			err := file.SaveToml(fs, initial, configFile)
			ExpectNoErr(err)

			testfileSymlink = symlink.Symlink{
				Target: path.UnexpandHome(config.Dotfiles+"/testfile", config.UserHome),
				Link:   path.UnexpandHome(config.UserHome+"/testfile", config.UserHome),
			}

			expectNewSymlink = func(expected symlink.Symlink) {
				var actual symlink.Config
				err := file.ReadToml(fs, &actual, configFile)
				ExpectNoErr(err)

				Expect(actual.Symlinks).Should(Equal(append(initial.Symlinks, expected)))
			}
		})

		It("should append the added symlink", func() {
			_, err := mgr.Add(testfile, "")
			Expect(err).To(Succeed())

			expectNewSymlink(testfileSymlink)
		})

		It("should not add mulitple entries for the same symlink", func() {
			_, err := mgr.Add(testfile, "")
			Expect(err).To(Succeed())
			_, err = mgr.Add(testfile, "")
			Expect(err).To(Succeed())

			expectNewSymlink(testfileSymlink)
		})
	})
})

func ExpectNoErr(err error) {
	if err != nil {
		Expect(err).To(BeNil())
	}
}
