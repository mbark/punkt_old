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
	base := packr.NewBox("./template")
	err := base.Walk(copyAll)
	if err != nil {
		logrus.WithError(err).Error("Unable to unpack ansible directories")
	}
}

func copyAll(src string, file packr.File) error {
	logrus.WithFields(logrus.Fields{
		"path": src,
	}).Info("Copying file")

	dest := "./" + src
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

	err = newFile.Sync()
	if err != nil {
		return err
	}

	return nil
}

// SaveStruct saves the given interface as yaml in the correct place
func SaveStruct(role string, content interface{}) bool {
	s := saverFromStruct(role, content)
	return s.Save()
}

// SaveYaml saves the given yaml for a role to the correct place
func SaveYaml(role string, content []byte) bool {
	s := newSaver(role, content)
	return s.Save()
}

type saver struct {
	role    string
	file    string
	content []byte
	logger  *logrus.Entry
}

func saverFromStruct(role string, content interface{}) *saver {
	out, err := yaml.Marshal(&content)
	if err != nil {
		logrus.WithField("role", role).WithError(err).Error("Unable to marshal db to yaml")
		return nil
	}

	return newSaver(role, out)
}

func newSaver(role string, content []byte) *saver {
	file := "roles/" + role + "/tasks/main.yml"
	logger := logrus.WithFields(logrus.Fields{
		"role": role,
		"file": file,
	})

	return &saver{
		role:    role,
		file:    file,
		content: content,
		logger:  logger,
	}
}

func (s saver) Save() bool {
	err := path.CreateNecessaryDirectories(s.file)
	if err != nil {
		s.logger.WithError(err).Error("Unable to create necessary directories")
		return false
	}

	f, err := os.Create(s.file)
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

	s.logger.Info("Succesfully wrote to backend database file")
	return true
}
