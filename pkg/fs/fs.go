package fs

import (
	"os"
	"os/user"

	billy "gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/osfs"
)

// Snapshot describes a snapshot of the Filesystem, Fs. The WorkingDir
// and UserHome are set initially and can be re-used.
type Snapshot struct {
	Fs         billy.Filesystem
	WorkingDir string
	UserHome   string
}

// NewSnapshot takes a snapshot of the current filesystem, saving the
// user's home directory and working dir.
func NewSnapshot() (*Snapshot, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return &Snapshot{
		Fs:         osfs.New("/"),
		WorkingDir: cwd,
		UserHome:   usr.HomeDir,
	}, nil
}
