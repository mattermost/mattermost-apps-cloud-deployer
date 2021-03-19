package exechelper

import (
	"os/exec"

	log "github.com/sirupsen/logrus"
)

func RemoveLocalFiles(files []string, logger log.FieldLogger) error {
	for _, file := range files {
		logger.Infof("Removing file %s", file)
		cmd := exec.Command("rm", "-rf", file)
		_, err := cmd.Output()
		if err != nil {
			return err
		}
	}

	return nil
}
