package symlink_test

import (
	"os"
	"path/filepath"

	g "github.com/onsi/ginkgo"
	m "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"

	"github.com/mbark/punkt/conf"
	"github.com/mbark/punkt/file"
	"github.com/mbark/punkt/mgr/symlink"
	"github.com/mbark/punkt/path"
)

var _ = g.Describe("Symlink: Add", func() {
	var fs billy.Filesystem
	var config conf.Config
	var mgr *symlink.Manager
	var testfile string

	g.BeforeEach(func() {
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

	g.It("should give an error for a non-existant file", func() {
		_, err := mgr.Add("nonExistantFile")
		m.Expect(err).NotTo(m.Succeed())
	})

	g.It("should be possible to create a symlink", func() {
		_, err := mgr.Add(testfile)
		m.Expect(err).To(m.Succeed())

		info, err := fs.Lstat(testfile)
		ExpectNoErr(err)
		m.Expect(info.Mode() & os.ModeSymlink).NotTo(m.Equal(0))

		f, err := fs.Readlink(testfile)
		ExpectNoErr(err)
		m.Expect(f).To(m.Equal(config.Dotfiles + "/testfile"))
	})

	g.It("should fail to add a file if the new location for it already exists", func() {
		_, err := fs.Create(config.Dotfiles + "/testfile")
		ExpectNoErr(err)

		_, err = mgr.Add(testfile)
		m.Expect(err).NotTo(m.Succeed())
	})

	g.Context("when adding symlinks", func() {
		var link string

		g.BeforeEach(func() {
			link = filepath.Join(config.UserHome, "link")
			err := fs.Symlink(testfile, link)
			m.Expect(err).To(m.BeNil())
		})

		g.It("should say it exists", func() {
			s, err := mgr.Add(link)
			m.Expect(err).To(m.Succeed())

			m.Expect(s.Exists()).To(m.BeTrue())
		})

	})

	g.Context("when reading and saving to the symlinks.yml file", func() {
		var initial []symlink.Symlink
		var testfileSymlink symlink.Symlink

		g.BeforeEach(func() {
			initial = []symlink.Symlink{
				{
					From: "~/some/file",
					To:   "~/to/some/file",
				},
			}
			err := file.SaveYaml(fs, initial, config.Dotfiles, "symlinks")
			ExpectNoErr(err)

			testfileSymlink = symlink.Symlink{
				From: path.UnexpandHome(config.Dotfiles+"/testfile", config.UserHome),
				To:   path.UnexpandHome(config.UserHome+"/testfile", config.UserHome),
			}
		})

		g.It("should append the added symlink", func() {
			_, err := mgr.Add(testfile)
			m.Expect(err).To(m.Succeed())

			actual := []symlink.Symlink{}
			err = file.Read(fs, &actual, config.Dotfiles, "symlinks")
			ExpectNoErr(err)

			m.Expect(actual).Should(m.Equal(append(initial, testfileSymlink)))
		})

		g.It("should not add mulitple entries for the same symlink", func() {
			_, err := mgr.Add(testfile)
			m.Expect(err).To(m.Succeed())
			_, err = mgr.Add(testfile)
			m.Expect(err).To(m.Succeed())

			actual := []symlink.Symlink{}
			err = file.Read(fs, &actual, config.Dotfiles, "symlinks")
			ExpectNoErr(err)

			m.Expect(actual).Should(m.Equal(append(initial, testfileSymlink)))
		})

		g.It("should not make a non-home relative path relative to home", func() {
			err := fs.MkdirAll("/dir", os.ModePerm)
			ExpectNoErr(err)

			_, err = fs.Create("/dir/absfile")
			ExpectNoErr(err)

			_, err = mgr.Add("/dir/absfile")
			m.Expect(err).To(m.Succeed())

			actual := []symlink.Symlink{}
			err = file.Read(fs, &actual, config.Dotfiles, "symlinks")
			ExpectNoErr(err)

			m.Expect(actual).Should(m.Equal(append(initial, symlink.Symlink{
				From: path.UnexpandHome(config.Dotfiles+"/dir/absfile", config.UserHome),
				To:   "/dir/absfile",
			})))
		})
	})
})

func ExpectNoErr(err error) {
	if err != nil {
		m.Expect(err).To(m.BeNil())
	}
}
