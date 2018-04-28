package git_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	billy "gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/osfs"
	"gopkg.in/src-d/go-billy.v4/util"
	goGit "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"

	"github.com/mbark/punkt/mgr/git"
)

func newStorage(fs billy.Filesystem, name string) (storage.Storer, billy.Filesystem) {
	worktree, err := fs.Chroot(name)
	Expect(err).To(BeNil())

	dotGit, err := worktree.Chroot(".git")
	Expect(err).To(BeNil())

	storage, err := filesystem.NewStorage(dotGit)
	Expect(err).To(BeNil())

	return storage, worktree
}

func openRepository(fs billy.Filesystem, name string) *goGit.Repository {
	storage, worktree := newStorage(fs, name)
	repo, err := goGit.Open(storage, worktree)
	Expect(err).To(BeNil())

	return repo
}

func newRepository(fs billy.Filesystem, name string, remote *config.RemoteConfig) (*goGit.Repository, billy.Filesystem) {
	storage, worktree := newStorage(fs, name)

	repo, err := goGit.Init(storage, worktree)
	Expect(err).To(BeNil())

	if remote != nil {
		_, err = repo.CreateRemote(remote)
	} else {
		_, err = repo.CreateRemote(&config.RemoteConfig{
			Name:  goGit.DefaultRemoteName,
			Fetch: []config.RefSpec{},
			URLs:  []string{"/some/path"},
		})
	}
	Expect(err).To(BeNil())

	return repo, worktree
}

func addCommit(repo *goGit.Repository) plumbing.Hash {
	w, err := repo.Worktree()
	Expect(err).To(BeNil())

	loc, err := time.LoadLocation("")
	Expect(err).To(BeNil())
	t := time.Date(2015, time.January, 1, 12, 59, 0, 0, loc)

	hash, err := w.Commit("msg", &goGit.CommitOptions{
		Author: &object.Signature{
			Name:  "John Doe",
			Email: "john@doe.org",
			When:  t,
		},
	})
	Expect(err).To(BeNil())

	return hash
}

var _ = Describe("Git: Repo Manager", func() {
	var fs billy.Filesystem
	var tmpdir string
	var mgr git.RepoManager

	BeforeEach(func() {
		logrus.SetLevel(logrus.PanicLevel)
		dir, err := ioutil.TempDir("", "git-repo-mgr")
		Expect(err).To(BeNil())
		tmpdir = dir
		fs = osfs.New(tmpdir)

		mgr = git.NewRepoManager(fs)
	})

	AfterEach(func() {
		err := os.RemoveAll(tmpdir)
		Expect(err).To(BeNil())
	})

	It("should be possible to construct", func() {
		Expect(mgr).NotTo(BeNil())
	})

	Context("Dump", func() {
		It("should dump the config for the specified repo", func() {
			repo, _ := newRepository(fs, "repo", nil)
			actual, err := repo.Config()
			Expect(err).To(BeNil())

			expected, err := mgr.Dump("repo")
			Expect(err).To(BeNil())

			Expect(expected.Name).To(Equal("repo"))
			Expect(expected.Config).To(Equal(actual))
		})

		It("should fail if the directory doesn't exist", func() {
			expected, err := mgr.Dump("/some/path")

			Expect(err).NotTo(BeNil())
			Expect(expected).To(BeNil())
		})

		It("should fail if the directory isn't a repository", func() {
			err := fs.MkdirAll("/some/path", os.ModePerm)
			Expect(err).To(BeNil())
			expected, err := mgr.Dump("/some/path")

			Expect(err).NotTo(BeNil())
			Expect(expected).To(BeNil())
		})

		It("should fail if the storage can't be allocated", func() {
			expected, err := mgr.Dump("../../../")

			Expect(err).NotTo(BeNil())
			Expect(expected).To(BeNil())
		})
	})

	Context("Ensure", func() {
		It("should succeed if the repository already exists", func() {
			repoPath := filepath.Join(tmpdir, "repo")
			newRepository(fs, repoPath, nil)
			repo, err := mgr.Dump(repoPath)
			Expect(err).To(BeNil())

			Expect(mgr.Ensure(*repo)).To(Succeed())
		})

		It("should clone the repository if it doesn't exist", func() {
			name := filepath.Join(tmpdir, "repo")
			origin, path := newRepository(fs, "origin", nil)
			addCommit(origin)
			newRepository(fs, name, &config.RemoteConfig{
				Name: goGit.DefaultRemoteName,
				URLs: []string{path.Root()},
			})

			repo, err := mgr.Dump(name)
			Expect(err).To(BeNil())

			err = util.RemoveAll(fs, name)
			Expect(err).To(BeNil())

			Expect(mgr.Ensure(*repo)).To(Succeed())

			repository := openRepository(fs, name)
			Expect(repository).NotTo(BeNil())
			config, err := repository.Config()
			Expect(err).To(BeNil())
			Expect(config.Remotes).To(Equal(repo.Config.Remotes))
			Expect(config.Core).To(Equal(repo.Config.Core))
		})

		It("should fail if the repository can't be cloned", func() {
			name := "repo"
			newRepository(fs, name, nil)
			repo, err := mgr.Dump(name)
			Expect(err).To(BeNil())

			err = util.RemoveAll(fs, name)
			Expect(err).To(BeNil())

			Expect(mgr.Ensure(*repo)).NotTo(Succeed())
		})

		It("should fail if storage can't be allocated", func() {
			Expect(mgr.Ensure(git.Repo{
				Path: "../../",
			})).NotTo(Succeed())
		})
	})

	Context("Update", func() {
		var origin *goGit.Repository
		var repository *goGit.Repository

		BeforeEach(func() {
			r1, path := newRepository(fs, "origin", nil)
			r2, _ := newRepository(fs, "repo", &config.RemoteConfig{
				Name: goGit.DefaultRemoteName,
				URLs: []string{path.Root()},
			})

			origin = r1
			repository = r2
		})

		It("should update the repository", func() {
			hash := addCommit(origin)

			updated, err := mgr.Update("repo")
			Expect(err).To(BeNil())
			Expect(updated).To(BeTrue())

			commit, err := repository.CommitObject(hash)
			Expect(err).To(BeNil())
			Expect(commit).NotTo(BeNil())
		})

		It("should succeed if the repository is already up to date", func() {
			addCommit(origin)

			_, err := mgr.Update("repo")
			Expect(err).To(BeNil())
			updated, err := mgr.Update("repo")

			Expect(err).To(BeNil())
			Expect(updated).To(BeFalse())
		})

		It("should fail if the default remote doesn't exist", func() {
			repository, _ = newRepository(fs, "noRemote", nil)

			updated, err := mgr.Update("repo")

			Expect(updated).To(BeFalse())
			Expect(err).NotTo(BeNil())
		})

		It("should fail if the repository doesn't exist", func() {
			err := util.RemoveAll(fs, "repo")
			Expect(err).To(BeNil())

			updated, err := mgr.Update("repo")
			Expect(err).NotTo(BeNil())
			Expect(updated).To(BeFalse())
		})

		It("should fail if the repository's storage can't be created", func() {
			_, err := mgr.Update("../../")
			Expect(err).NotTo(BeNil())
		})
	})
})
