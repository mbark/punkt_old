package db

import (
	"io"
	"os"

	"github.com/mbark/punkt/path"

	"github.com/gobuffalo/packr"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// CreateStructure ...
func CreateStructure() {
	err := os.MkdirAll("./usr", os.ModePerm)
	if err != nil {
		logrus.WithError(err).Fatal("Unable to create directory to store usr configuration")
	}

	base := packr.NewBox("./template")
	err = base.Walk(copyAll)
	if err != nil {
		logrus.WithError(err).Fatal("Unable to unpack ansible directories")
	}
}

func copyAll(src string, file packr.File) error {
	logrus.WithFields(logrus.Fields{
		"path": src,
	}).Debug("Copying file")

	dest := "./ansible/" + src
	err := path.CreateNecessaryDirectories(dest)
	if err != nil {
		return err
	}

	newFile, err := os.Create(dest)
	if err != nil {
		return err
	}

	_, err = io.Copy(newFile, file)
	if err != nil {
		return err
	}

	return newFile.Sync()
}

// SaveStruct ...
func SaveStruct(path string, content interface{}) bool {
	out, err := yaml.Marshal(&content)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"role": path,
		}).WithError(err).Error("Unable to marshal db to yaml")
		return false
	}

	s := newSaver(path, out)
	return s.Save()
}

type saver struct {
	path    string
	content []byte
	logger  *logrus.Entry
}

func newSaver(path string, content []byte) *saver {
	logger := logrus.WithFields(logrus.Fields{
		"path": path,
	})

	return &saver{
		path:    path,
		content: content,
		logger:  logger,
	}
}

func (s saver) Save() bool {
	err := path.CreateNecessaryDirectories(s.path)
	if err != nil {
		s.logger.WithError(err).Error("Unable to create necessary directories")
		return false
	}

	f, err := os.Create(s.path)
	if err != nil {
		s.logger.WithError(err).Error("Unable to create file")
		return false
	}

	defer f.Close()

	_, err = f.Write(s.content)
	if err != nil {
		s.logger.WithError(err).Error("Unable to write to file")
		return false
	}

	f.Sync()

	s.logger.Info("Successfully wrote to backend database file")
	return true
}
