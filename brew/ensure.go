package brew

import ()

// Ensure ...
func Ensure() {
	bundle("--no-upgrade")
}
