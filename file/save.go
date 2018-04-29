package file

import (
	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-billy.v4"

	"github.com/mbark/punkt/path"
	"github.com/pkg/errors"
)

// SaveToml ...
func SaveToml(fs billy.Filesystem, content interface{}, file string) error {
	err := path.CreateNecessaryDirectories(fs, file)
	if err != nil {
		return err
	}

	f, err := fs.Create(file)
	if err != nil {
		return errors.Wrapf(err, "failed to create file [file: %s]", file)
	}

	encoder := toml.NewEncoder(f)
	return encoder.Encode(content)
}

// Save ...
func Save(fs billy.Filesystem, content string, file string) error {
	if content == "" {
		logrus.WithField("file", file).Info("no content to save to file, ignoring")
		return nil
	}

	err := path.CreateNecessaryDirectories(fs, file)
	if err != nil {
		return errors.Wrapf(err, "failed to create necessary directories [file: %s]", file)
	}

	f, err := fs.Create(file)
	if err != nil {
		return errors.Wrapf(err, "failed to create file [file: %s]", file)
	}

	defer f.Close()

	_, err = f.Write([]byte(content))
	return err
}
