package main

// RunConfig is a struct which contains the configuration for running, primarily
// set by the command line arguments
type RunConfig struct {
	DryRun     bool
	ConfigFile string
	Config     Config
}

// Config is ...
type Config struct {
	ParentDir    string
	Symlinks     map[string]string   `yaml:"symlinks"`
	BackendFiles map[string]string   `yaml:"backends"`
	Backends     map[string]Backend  `yaml:"-"`
	Tasks        []map[string]string `yaml:"tasks"`
}

// Backend is ...
type Backend struct {
	Name    string `yaml:"-"`
	List    string `yaml:"list"`
	Update  string `yaml:"update"`
	Install string `yaml:"install"`
}
