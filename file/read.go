package file

import (
	"bytes"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-billy.v4"
)

// ErrNoSuchFile is returned if the file to read from can't be found
var ErrNoSuchFile = errors.New("file doesn't exist")

func open(fs billy.Filesystem, path string) (*bytes.Buffer, error) {
	file, err := fs.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open file [file: %s]", path)
	}

	defer file.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(file)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read from file [file: %s]", path)
	}

	return buf, nil
}

// ReadToml the file in the given directory and marshal it to the given struct
func ReadToml(fs billy.Filesystem, out interface{}, file string) error {
	_, err := fs.Stat(file)
	if err != nil {
		return ErrNoSuchFile
	}

	buf, err := open(fs, file)
	if err != nil {
		return errors.Wrapf(err, "failed to open file [file: %s]", file)
	}

	err = toml.Unmarshal(buf.Bytes(), out)
	if err != nil {
		return errors.Wrapf(err, "failed to unmarshal file [file: %s]", file)
	}

	return nil
}
