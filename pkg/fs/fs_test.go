package fs_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/mbark/punkt/pkg/fs"
	"github.com/mbark/punkt/testmock"
)

func TestMgr(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fs Suite")
}

var _ = Describe("Fs Manager", func() {
	var snapshot fs.Snapshot

	BeforeEach(func() {
		snapshot, _ = testmock.Setup()
	})

	Context("CreateNecessaryDirectories", func() {
		It("should return an error if the directories can't be made", func() {
			_, err := snapshot.Fs.Create("/foo")
			Expect(err).To(BeNil())

			Expect(snapshot.CreateNecessaryDirectories("/foo/bar")).NotTo(Succeed())
		})
	})

	Context("ReadToml", func() {
		It("should return no such file if the file doesn't exit", func() {
			var out interface{}
			err := snapshot.ReadToml(&out, "/non/existent")

			Expect(err).To(Equal(fs.ErrNoSuchFile))
		})

		It("should return an error if the file can't be opened", func() {
			var out interface{}
			err := snapshot.Fs.MkdirAll("/foo", os.ModePerm)
			Expect(err).To(BeNil())

			Expect(snapshot.ReadToml(&out, "/foo")).NotTo(Succeed())
		})
	})

	Context("Save{,Toml}", func() {
		It("should fail to save if necessary directories can't be made", func() {
			_, err := snapshot.Fs.Create("/foo")
			Expect(err).To(BeNil())

			Expect(snapshot.SaveToml(make(map[string]string), "/foo/bar")).NotTo(Succeed())
			Expect(snapshot.Save("foo", "/foo/bar")).NotTo(Succeed())
		})

		It("should fail to save if the file can't be created", func() {
			err := snapshot.Fs.MkdirAll("/foo", os.ModePerm)
			Expect(err).To(BeNil())

			Expect(snapshot.SaveToml(make(map[string]string), "/foo")).NotTo(Succeed())
			Expect(snapshot.Save("foo", "/foo")).NotTo(Succeed())
		})
	})

})
