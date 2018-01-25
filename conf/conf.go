package conf

import ()

// Config ...
type Config struct {
	PunktHome string
	Dotfiles  string
	UserHome  string
}

func NewConfig(punktHome, dotfiles string, userHome string) *Config {
	return &Config{
		PunktHome: punktHome,
		Dotfiles:  dotfiles,
		UserHome:  userHome,
	}
}
