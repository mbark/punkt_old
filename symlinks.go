package main

import (
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

// CreateSymlinks will construct all symlinks specified in the configuration
// file. Returns true if all symlinks were successfully created, otherwise
// false. It will attempt to create all symlinks, even if one fails.
func CreateSymlinks() bool {
	hadError := false

	for to, from := range runConfig.Config.Symlinks {
		log.WithFields(log.Fields{
			"to":   to,
			"from": from,
		}).Info("Creating symlink")

		from = filepath.Join(runConfig.Config.ParentDir, from)
		to = filepath.Join(runConfig.Config.ParentDir, to)

		_, err := os.Stat(from)
		if err != nil {
			log.WithError(err).WithField("from", RelPath(from)).Warning("No such file")
			hadError = true
			continue
		}

		if runConfig.DryRun {
			continue
		}

		err = CreateNecessaryDirectories(to)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"from": RelPath(from),
				"to":   RelPath(to),
			}).Warning("Unable to create necessary directories")
			hadError = true

			continue
		}

		err = os.Symlink(from, to)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"from": RelPath(from),
				"to":   RelPath(to),
			}).Warning("Unable to create symlink")
			hadError = true
		}
	}

	return hadError
}
