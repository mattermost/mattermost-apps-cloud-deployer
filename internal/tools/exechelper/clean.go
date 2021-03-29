package exechelper

import (
	"os/exec"

	log "github.com/sirupsen/logrus"
)

// RemoveLocalFiles is used to clean local files after being processed.
func RemoveLocalFiles(files []string, logger log.FieldLogger) error {
	for _, file := range files {
		logger.Infof("Removing file %s", file)
		cmd := exec.Command("rm", "-rf", file)
		if _, err := cmd.Output(); err != nil {
			return err
		}
	}

	return nil
}
