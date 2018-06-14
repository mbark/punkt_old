package testmock

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-billy.v4/memfs"

	"github.com/mbark/punkt/pkg/conf"
	"github.com/mbark/punkt/pkg/fs"
	"github.com/mbark/punkt/pkg/printer"
)

// Setup can be used by tests to mock out some things that
// should typically be done, such as making sure no commands or
// the printer don't print output to stdout and so on.
func Setup() (fs.Snapshot, conf.Config) {
	logrus.SetLevel(logrus.PanicLevel)

	snapshot := fs.Snapshot{Fs: memfs.New(), UserHome: "/home", WorkingDir: "/home/path"}
	config := conf.Config{
		PunktHome: "/home/.config/punkt",
		Dotfiles:  "/home/.dotfiles",
	}

	NoOutput()
	printer.Log.Out = ioutil.Discard

	return snapshot, config
}
