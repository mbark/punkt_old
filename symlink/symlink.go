package symlink

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
	dir      string
	depth    int
	dotfiles string
	Symlinks []symlink
}

type symlink struct {
	From string
	To   string
}

// Dump ...
func Dump(directories []string, depth int, dest, from string) {
	symlinks := find(directories, depth, from)
	logrus.WithFields(logrus.Fields{
		"symlinks":    symlinks,
		"directories": directories,
		"depth":       depth,
	}).Debug("Found the following symlinks")

	db.SaveVars("symlinks", symlinks, dest)
}

func find(directories []string, depth int, from string) []symlink {
	var symlinks []symlink

	for _, dir := range directories {
		f := finder{
			dir:      path.ExpandHome(dir),
			depth:    depth,
			dotfiles: from,
			Symlinks: []symlink{},
		}

		logrus.WithFields(logrus.Fields{
			"finder": f,
		}).Debug("Searching for symlinks")

		filepath.Walk(f.dir, f.walkFunc)
		symlinks = append(symlinks, f.Symlinks...)
	}

	return symlinks
}

func (f *finder) walkFunc(currpath string, info os.FileInfo, err error) error {
	if currpath == f.dir {
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

		if strings.HasPrefix(to, f.dotfiles) {
			f.Symlinks = append(f.Symlinks, symlink{
				To:   path.UnexpandHome(currpath),
				From: path.UnexpandHome(to),
			})
		}
	}

	for _, val := range blacklist {
		if strings.HasSuffix(currpath, val) {
			return filepath.SkipDir
		}
	}

	if info.IsDir() {
		rel, err := filepath.Rel(f.dir, currpath)
		if err != nil {
			return err
		}

		if strings.Count(rel, "/") >= f.depth {
			return filepath.SkipDir

		}
	}

	return nil
}
