package symlink

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mbark/punkt/path"

	"github.com/gobwas/glob"
	"github.com/sirupsen/logrus"
)

var blacklist = []string{"*/.Trash", "*/.git"}

type finder struct {
	dir      string
	depth    int
	Symlinks []Symlink
	ignore   []glob.Glob
}

// Dump ...
func Dump(directories []string, depth int, ignore []string) []Symlink {
	symlinks := find(directories, depth, ignore)
	logrus.WithFields(logrus.Fields{
		"symlinks":    symlinks,
		"directories": directories,
		"depth":       depth,
	}).Debug("Found the following symlinks")

	return symlinks
}

func find(directories []string, depth int, ignore []string) []Symlink {
	var symlinks []Symlink

	for _, dir := range directories {
		f := finder{
			dir:      path.ExpandHome(dir),
			depth:    depth,
			Symlinks: []Symlink{},
			ignore:   constructIgnoreGlobs(ignore),
		}

		logrus.WithFields(logrus.Fields{
			"finder": f,
		}).Debug("Searching for symlinks")

		filepath.Walk(f.dir, f.walkFunc)
		symlinks = append(symlinks, f.Symlinks...)
	}

	return symlinks
}

func constructIgnoreGlobs(ignore []string) []glob.Glob {
	globs := []glob.Glob{}
	for _, ignored := range append(blacklist, ignore...) {
		glob, err := glob.Compile(ignored)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"pattern": ignored,
			}).WithError(err).Warn("Unable to compile pattern to glob, ignoring")
			continue
		}

		globs = append(globs, glob)
	}

	return globs
}

func (f *finder) walkFunc(currpath string, info os.FileInfo, err error) error {
	if currpath == f.dir {
		return nil
	}

	if err != nil {
		return nil
	}

	if f.isBlacklisted(currpath) {
		return filepath.SkipDir
	}

	if info.Mode()&os.ModeSymlink != 0 {
		to, err := filepath.EvalSymlinks(currpath)
		if err != nil {
			return err
		}

		_, err = os.Stat(to)
		if err != nil {
			return err
		}

		f.Symlinks = append(f.Symlinks, Symlink{
			To:   path.UnexpandHome(currpath),
			From: path.UnexpandHome(to),
		})

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

func (f finder) isBlacklisted(path string) bool {
	for _, ignored := range f.ignore {
		if ignored.Match(path) {
			return true
		}
	}

	return false
}
