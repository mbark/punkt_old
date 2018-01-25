package file

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/mbark/punkt/path"
)

// SaveYaml the given interface to the given directory with the specified name,
// the suffix is added by defautl
func SaveYaml(content interface{}, dest, name string) error {
	out, err := yaml.Marshal(&content)
	if err != nil {
		logrus.WithError(err).Error("Unable to marshal db to yaml")
		return err
	}

	path := filepath.Join(dest, name+".yml")
	s := newSaver(path, out)
	return s.Save()
}

// Save ...
func Save(content string, dest, name string) error {
	path := filepath.Join(dest, name)

	logrus.WithFields(logrus.Fields{
		"content": content,
		"path":    path,
	}).Debug("Saving content to file")
	s := newSaver(path, []byte(content))
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

func (s saver) Save() error {
	err := path.CreateNecessaryDirectories(s.path)
	if err != nil {
		s.logger.WithError(err).Error("Unable to create necessary directories")
		return err
	}

	f, err := os.Create(s.path)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(s.content)
	if err != nil {
		return err
	}

	err = f.Sync()
	if err != nil {
		return err
	}

	s.logger.Info("Successfully wrote to backend database file")
	return nil
}
