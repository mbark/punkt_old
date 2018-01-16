package symlink

import (
	"github.com/sirupsen/logrus"
)

// Ensure goes through the list of symlinks ensuring they exist
func Ensure(symlinks []Symlink) {
	for _, symlink := range symlinks {
		s := symlink.expand()

		if s.Exists() {
			logrus.WithFields(logrus.Fields{
				"from": s.From,
				"to":   s.To,
			}).Debug("Symlink already exists, not creating")
		} else {
			s.Create()
		}
	}
}
