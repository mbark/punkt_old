package file

import (
	"github.com/BurntSushi/toml"
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
func Save(fs billy.Filesystem, content string, file string) error {
	if content == "" {
		return nil
	}

	err := path.CreateNecessaryDirectories(fs, file)
	if err != nil {
		return err
	}

	f, err := fs.Create(file)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write([]byte(content))
	return err
}
