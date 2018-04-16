package file

import (
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-billy.v4"

	"github.com/mbark/punkt/path"
)

// SaveToml ...
func SaveToml(fs billy.Filesystem, content interface{}, file string) error {
	err := path.CreateNecessaryDirectories(fs, file)
	if err != nil {
		return err
	}

	f, err := fs.Create(file)
	if err != nil {
		return err
	}

	encoder := toml.NewEncoder(f)
	return encoder.Encode(content)
}

// Save ...
func Save(fs billy.Filesystem, content string, dest, name string) error {
	if content == "" {
		return nil
	}

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
