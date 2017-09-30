package main

type RunConfig struct {
	DryRun     bool
	ConfigFile string
	Config     Config
}

type Config struct {
	ParentDir string
	Symlinks  map[string]string   `yaml:"symlinks"`
	Backends  map[string]string   `yaml:"backends"`
	Tasks     []map[string]string `yaml:"tasks"`
}

type Backend struct {
	Bootstrap string `yaml:"bootstrap"`
	List      string `yaml:"list"`
	Update    string `yaml:"update"`
	Install   string `yaml:"install"`
}
