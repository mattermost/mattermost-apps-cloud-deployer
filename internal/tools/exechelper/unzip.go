package exechelper

import (
	"fmt"
	"os/exec"
	"strings"
)

// UnzipBundle is used to unzip downloaded bundles.
func UnzipBundle(dir, bundle string) error {
	bundleName := strings.TrimSuffix(bundle, ".zip")

	cmdMkdir := exec.Command("mkdir", fmt.Sprintf("%s/%s", dir, bundleName))
	if _, err := cmdMkdir.Output(); err != nil {
		return err
	}

	cmdUnzip := exec.Command("unzip", fmt.Sprintf("%s/%s", dir, bundle), "-d", fmt.Sprintf("%s/%s", dir, bundleName))
	if _, err := cmdUnzip.Output(); err != nil {
		return err
	}
	return nil
}
