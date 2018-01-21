package brew

import ()

// Ensure ...
func Ensure(dotfiles string) {
	bundle("--no-upgrade")
}
