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
)

func main() {
	kingpin.Parse()

	filename, err := filepath.Abs(*config_file)
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
	fmt.Printf("%s\n", config)

	hadError := false

	for to, from := range config.Symlinks {
		from = filepath.Join(config.ParentDir, from)
		to = filepath.Join(config.ParentDir, to)
		createIntermittentDirectories(to)
		fmt.Printf("ln -s %s %s\n", from, to)
		err = os.Symlink(from, to)
		if err != nil {
			fmt.Printf("Unable to create symlink %v\n", err)
			hadError = true
		}
	}

	if hadError {
		os.Exit(1)
	}
}

func createIntermittentDirectories(file string) {
	dir := filepath.Dir(file)
	os.MkdirAll(dir, os.ModePerm)
}
