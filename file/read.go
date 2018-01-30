package file

import (
	"bytes"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/yaml.v2"
)

// Read the file in the given directory and marshal it to the given struct
func Read(fs billy.Filesystem, out interface{}, dest, name string) error {
	path := filepath.Join(dest, name+".yml")
	logger := logrus.WithFields(logrus.Fields{
		"file": path,
	})

	file, err := fs.Open(path)
	if err != nil {
		return err
	}

	defer file.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(file)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(buf.Bytes(), out)
	if err != nil {
		logger.WithError(err).Error("Unable to unmarshal file to yaml")
		return err
	}

	return nil
}
