package exechelper

import (
	"os/exec"
	"path"
	"strings"
)

// UnzipBundle is used to unzip downloaded bundles.
func UnzipBundle(dir, bundle string) error {
	bundleName := strings.TrimSuffix(bundle, ".zip")

	cmdMkdir := exec.Command("mkdir", path.Join(dir, bundleName))
	if _, err := cmdMkdir.Output(); err != nil {
		return err
	}

	cmdUnzip := exec.Command("unzip", path.Join(dir, bundle), "-d", path.Join(dir, bundleName))
	if _, err := cmdUnzip.Output(); err != nil {
		return err
	}
	return nil
}
