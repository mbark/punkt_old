package file

import (
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/yaml.v2"

	"github.com/mbark/punkt/path"
)

// SaveYaml the given interface to the given directory with the specified name,
// the suffix is added by default
func SaveYaml(fs billy.Filesystem, content interface{}, dest, name string) error {
	out, err := yaml.Marshal(&content)
	if err != nil {
		logrus.WithError(err).Error("Unable to marshal db to yaml")
		return err
	}
	logrus.WithField("out", out).Debug("marshalled")

	path := filepath.Join(dest, name+".yml")
	s := newSaver(fs, path, out)
	return s.Save()
}

// Save ...
func Save(fs billy.Filesystem, content string, dest, name string) error {
	path := filepath.Join(dest, name)

	logrus.WithFields(logrus.Fields{
		"content": content,
		"path":    path,
	}).Debug("Saving content to file")
	s := newSaver(fs, path, []byte(content))
	return s.Save()
}

type saver struct {
	fs      billy.Filesystem
	path    string
	content []byte
	logger  *logrus.Entry
}

func newSaver(fs billy.Filesystem, path string, content []byte) *saver {
	logger := logrus.WithFields(logrus.Fields{
		"path": path,
	})

	return &saver{
		fs:      fs,
		path:    path,
		content: content,
		logger:  logger,
	}
}

func (s saver) Save() error {
	err := path.CreateNecessaryDirectories(s.fs, s.path)
	if err != nil {
		return err
	}

	f, err := s.fs.Create(s.path)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(s.content)
	if err != nil {
		return err
	}

	s.logger.Info("Successfully wrote to backend database file")
	return nil
}
