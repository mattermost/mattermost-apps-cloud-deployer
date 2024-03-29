package terraform

import (
	"os/exec"

	appsutils "github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/pkg/errors"
)

// Cmd is the terraform command to execute.
type Cmd struct {
	terraformPath     string
	dir               string
	remoteStateBucket string
	logger            appsutils.Logger
}

// New creates a new instance of Cmd through which to execute terraform.
func New(dir, remoteStateBucket string, logger appsutils.Logger) (*Cmd, error) {
	if remoteStateBucket == "" {
		return nil, errors.New("remote state bucket cannot be an empty value")
	}
	terraformPath, err := exec.LookPath("terraform")
	if err != nil {
		return nil, errors.Wrap(err, "failed to find terraform installed on your PATH")
	}

	return &Cmd{
		terraformPath:     terraformPath,
		dir:               dir,
		remoteStateBucket: remoteStateBucket,
		logger:            logger,
	}, nil
}

// GetWorkingDirectory returns the working directory used by terraform.
func (c *Cmd) GetWorkingDirectory() string {
	return c.dir
}

// Close is a no-op.
func (c *Cmd) Close() error {
	return nil
}
