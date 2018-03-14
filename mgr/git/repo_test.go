package git_test

import (
	"testing"

	"github.com/mbark/punkt/mgr/git"

	g "github.com/onsi/ginkgo"
	m "github.com/onsi/gomega"
	billy "gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	goGit "gopkg.in/src-d/go-git.v4"
	goConfig "gopkg.in/src-d/go-git.v4/config"
	goStorage "gopkg.in/src-d/go-git.v4/storage/filesystem"
)

func TestGitRepo(t *testing.T) {
	m.RegisterFailHandler(g.Fail)
	g.RunSpecs(t, "Git Repo Suite")
}

var _ = g.Describe("Git: Repo", func() {
	var dest string
	var worktree billy.Filesystem

	g.BeforeEach(func() {
		fs := memfs.New()
		dest = "/"

		createGitRepo(fs, dest, "repo", nil)
		w, err := fs.Chroot(dest + "/repo")
		m.Expect(err).To(m.BeNil())

		worktree = w
	})

	g.It("should be possible to create a new repo", func() {
		repo, err := git.NewRepo(worktree, "")
		m.Expect(err).To(m.BeNil())
		m.Expect(repo).NotTo(m.BeNil())
	})

	g.It("should fail if there is no repo", func() {
		repo, err := git.NewRepo(memfs.New(), "notRepo")
		m.Expect(err).NotTo(m.BeNil())
		m.Expect(repo).To(m.BeNil())
	})

	g.It("should derive the name for the repo from the origin if necessary", func() {
		repo, err := git.NewRepo(worktree, "")
		m.Expect(err).To(m.BeNil())
		m.Expect(repo).NotTo(m.BeNil())
		m.Expect(repo.Name).To(m.Equal("repo"))
	})

	g.It("should use the given name if specified", func() {
		repo, err := git.NewRepo(worktree, "aname")
		m.Expect(err).To(m.BeNil())
		m.Expect(repo).NotTo(m.BeNil())
		m.Expect(repo.Name).To(m.Equal("aname"))
	})

	g.It("should use directory name if there is no default remote", func() {
		fs := memfs.New()
		createGitRepo(fs, "dir", "dirName", &goConfig.RemoteConfig{
			Name:  "nonDefault",
			Fetch: []goConfig.RefSpec{},
			URLs:  []string{"/path"},
		})

		w, err := fs.Chroot("/dir/dirName")
		m.Expect(err).To(m.BeNil())
		repo, err := git.NewRepo(w, "")
		m.Expect(err).To(m.BeNil())
		m.Expect(repo).NotTo(m.BeNil())
		m.Expect(repo.Name).To(m.Equal("dirName"))
	})
})

func createGitRepo(fs billy.Filesystem, dest, name string, remote *goConfig.RemoteConfig) {
	dir, err := fs.Chroot(dest + "/" + name)
	m.Expect(err).To(m.BeNil())

	storage, err := goStorage.NewStorage(dir)
	m.Expect(err).To(m.BeNil())
	repo, err := goGit.Init(storage, dir)
	m.Expect(err).To(m.BeNil())

	if remote != nil {
		_, err = repo.CreateRemote(remote)
	} else {
		_, err = repo.CreateRemote(&goConfig.RemoteConfig{
			Name:  "origin",
			Fetch: []goConfig.RefSpec{},
			URLs:  []string{"/some/path/" + name},
		})
	}
	m.Expect(err).To(m.BeNil())
}
