package fs

import (
	"io"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-billy.v4"
)

// ErrNoSuchFile is returned if the file to read from can't be found
var ErrNoSuchFile = errors.New("file doesn't exist")

// ReadToml the file in the given directory and marshal it to the given struct
func (snapshot Snapshot) ReadToml(out interface{}, file string) error {
	_, err := snapshot.Fs.Stat(file)
	if err != nil {
		return ErrNoSuchFile
	}

	f, err := snapshot.Fs.Open(file)
	if err != nil {
		return errors.Wrapf(err, "failed to open %s", file)
	}

	defer close(f)

	_, err = toml.DecodeReader(f, out)
	if err != nil {
		return errors.Wrapf(err, "failed to unmarshal %s", file)
	}

	return nil
}

// SaveToml ...
func (snapshot Snapshot) SaveToml(content interface{}, file string) error {
	f, err := snapshot.ensureFile(file)
	if err != nil {
		return err
	}

	defer close(f)

	encoder := toml.NewEncoder(f)
	return encoder.Encode(content)
}

// Save ...
func (snapshot Snapshot) Save(content string, file string) error {
	if content == "" {
		logrus.WithField("file", file).Info("no content to save to file, ignoring")
		return nil
	}

	f, err := snapshot.ensureFile(file)
	if err != nil {
		return err
	}

	defer close(f)

	_, err = f.Write([]byte(content))
	return err
}

func (snapshot Snapshot) ensureFile(file string) (billy.File, error) {
	err := snapshot.CreateNecessaryDirectories(file)
	if err != nil {
		return nil, err
	}

	f, err := snapshot.Fs.Create(file)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create %s", file)
	}

	return f, err
}

func close(c io.Closer) {
	err := c.Close()
	if err != nil {
		logrus.WithError(err).Error("error when closing file")
	}
}
