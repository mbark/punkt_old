package main

import (
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var (
	config_file = kingpin.Arg("config", "Configuration file").Required().String()
	dry_run     = kingpin.Flag("dryrun", "Just print what would have been done").Short('n').Bool()
)

func main() {
	kingpin.Parse()

	config := RunConfig{
		DryRun:     *dry_run,
		ConfigFile: *config_file,
	}

	config.Config = readConfigFile(*config_file)
	fmt.Printf("%s\n", config)

	var hadError bool = createSymlinks(config)

	if hadError {
		os.Exit(1)
	}
}

func readConfigFile(file string) Config {
	filename, err := filepath.Abs(file)
	if err != nil {
		log.Fatal(err)
	}
	_, err = os.Stat(filename)
	if os.IsNotExist(err) {
		log.Fatal(err)
	}

	contents, err := ioutil.ReadFile(*config_file)
	if err != nil {
		log.Fatal(err)
	}

	config := Config{}
	config.ParentDir = filepath.Dir(filename)
	err = yaml.Unmarshal(contents, &config)

	if err != nil {
		log.Fatal(err)
	}

	return config
}

func createSymlinks(config RunConfig) bool {
	hadError := false

	for to, from := range config.Config.Symlinks {
		from = filepath.Join(config.Config.ParentDir, from)
		to = filepath.Join(config.Config.ParentDir, to)
		fmt.Printf("ln -s %s %s\n", from, to)

		if !config.DryRun {
			err := createNecessaryDirectories(to)
			if err != nil {
				fmt.Printf("Unable to create necessary directories %s\n", err)
				hadError = true
			}

			err = os.Symlink(from, to)
			if err != nil {
				fmt.Printf("Unable to create symlink %v\n", err)
				hadError = true
			}
		}
	}

	return hadError
}

func createNecessaryDirectories(file string) error {
	dir := filepath.Dir(file)
	return os.MkdirAll(dir, os.ModePerm)
}
