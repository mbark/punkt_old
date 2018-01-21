package file

import (
	"io/ioutil"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Read the file in the given directory and marshal it to the given struct
func Read(out interface{}, dest, name string) {
	path := filepath.Join(dest, name+".yml")
	logger := logrus.WithFields(logrus.Fields{
		"file": path,
	})

	in, err := ioutil.ReadFile(path)
	if err != nil {
		logger.WithError(err).Fatal("Unable to read file")
	}

	err = yaml.Unmarshal(in, out)
	if err != nil {
		logger.WithError(err).Fatal("Unable to unmarshal file to yaml")
	}
}
