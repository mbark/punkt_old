package file

import (
	"bytes"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/yaml.v2"
)

func open(fs billy.Filesystem, dest, name string) (*bytes.Buffer, error) {
	path := filepath.Join(dest, name)
	logger := logrus.WithFields(logrus.Fields{
		"file": path,
	})

	file, err := fs.Open(path)
	if err != nil {
		logger.WithError(err).Warning("Unable to open file")
		return nil, err
	}

	defer file.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(file)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// Read the file in the given directory and marshal it to the given struct
func Read(fs billy.Filesystem, out interface{}, dest, name string) error {
	buf, err := open(fs, dest, name+".yml")
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(buf.Bytes(), out)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"dir":  dest,
			"name": name,
		}).WithError(err).Error("Unable to unmarshal file to yaml")
		return err
	}

	return nil
}

// ReadToml the file in the given directory and marshal it to the given struct
func ReadToml(fs billy.Filesystem, out interface{}, dest, name string) error {
	buf, err := open(fs, dest, name+".toml")
	if err != nil {
		return err
	}

	err = toml.Unmarshal(buf.Bytes(), out)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"dir":  dest,
			"name": name,
		}).WithError(err).Error("Unable to unmarshal file to yaml")
		return err
	}

	return nil
}
