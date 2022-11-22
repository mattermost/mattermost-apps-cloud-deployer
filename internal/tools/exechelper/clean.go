package exechelper

import (
	"os/exec"

	appsutils "github.com/mattermost/mattermost-plugin-apps/utils"
)

// RemoveLocalFiles is used to clean local files after being processed.
func RemoveLocalFiles(files []string, logger appsutils.Logger) error {
	for _, file := range files {
		logger.Infof("Removing file %s", file)
		cmd := exec.Command("rm", "-rf", file)
		if _, err := cmd.Output(); err != nil {
			return err
		}
	}

	return nil
}
