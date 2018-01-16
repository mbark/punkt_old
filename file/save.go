package file

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/mbark/punkt/path"
)

// Save the given interface to the given directory with the specified name,
// the suffix is added by defautl
func Save(content interface{}, dest, name string) bool {
	path := filepath.Join(dest, name+".yml")
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
