package file

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mbark/punkt/db"
	"github.com/mbark/punkt/path"

	"github.com/sirupsen/logrus"
)

var blacklist = []string{".Trash", ".git"}

type finder struct {
	home     string
	depth    int
	Symlinks []symlink
}

type symlink struct {
	From string
	To   string
}

// Dump ...
func Dump() {
	s := finder{
		depth: 2,
		home:  path.GetUserHome(),
	}

	s.find()
	logrus.WithFields(logrus.Fields{
		"symlinks": s.Symlinks,
		"depth":    s.depth,
	}).Debug("Found the following symlinks")

	db.SaveStruct("vars/symlinks.yml", s)

}

func (f *finder) find() {
	filepath.Walk(f.home, f.walkFunc)

	var filtered []symlink
	for _, link := range f.Symlinks {
		if strings.HasPrefix(link.From, f.home+"/dotfiles") {
			filtered = append(filtered, symlink{
				From: f.rewriteHome(link.From),
				To:   f.rewriteHome(link.To),
			})
		}
	}

	f.Symlinks = filtered
}

func (f *finder) rewriteHome(path string) string {
	return strings.Replace(path, f.home, "~", 1)
}

func (f *finder) walkFunc(currpath string, info os.FileInfo, err error) error {
	if currpath == f.home {
		return nil
	}

	if err != nil {
		return nil
	}

	if info.Mode()&os.ModeSymlink != 0 {
		to, err := os.Readlink(currpath)
		if err != nil {
			return err
		}

		f.Symlinks = append(f.Symlinks, symlink{
			To:   currpath,
			From: to,
		})
	}

	for _, val := range blacklist {
		if strings.HasSuffix(currpath, val) {
			return filepath.SkipDir
		}
	}

	if info.IsDir() {
		rel, err := filepath.Rel(f.home, currpath)
		if err != nil {
			return err
		}

		if strings.Count(rel, "/") >= f.depth {
			return filepath.SkipDir

		}
	}

	return nil
}
