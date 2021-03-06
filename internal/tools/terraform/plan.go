package terraform

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	model "github.com/mattermost/mattermost-apps/model"
	"github.com/pkg/errors"
)

type terraformOutput struct {
	Sensitive bool        `json:"sensitive"`
	Type      string      `json:"type"`
	Value     interface{} `json:"value"`
}

// Init invokes terraform init.
func (c *Cmd) Init(remoteKey string) error {
	_, _, err := c.run(
		"init",
		arg("backend-config", fmt.Sprintf("bucket=%s", c.remoteStateBucket)),
		arg("backend-config", fmt.Sprintf("key=%s", remoteKey)),
		arg("backend-config", "region=us-east-1"),
	)
	if err != nil {
		return errors.Wrap(err, "failed to invoke terraform init")
	}

	return nil
}

// Plan invokes terraform Plan.
func (c *Cmd) Plan(function model.Function) error {
	_, _, err := c.run(
		"plan",
		arg("input", "false"),
		arg("var", fmt.Sprintf("lambda_name=%s", function.Name)),
		arg("var", fmt.Sprintf("lambda_file=%s", function.ZipFile)),
		arg("var", fmt.Sprintf("environment=%s", function.Environment)),
		arg("var", fmt.Sprintf("bundle_name=%s", function.BundleName)),
		arg("var", fmt.Sprintf("handler=%s", function.Handler)),
		arg("var", fmt.Sprintf("runtime=%s", function.Runtime)),
		arg("var", fmt.Sprintf("private_subnet_ids=%s", os.Getenv("PrivateSubnetIDs"))),
	)
	if err != nil {
		return errors.Wrap(err, "failed to invoke terraform plan")
	}

	return nil
}

// Apply invokes terraform apply.
func (c *Cmd) Apply(function model.Function) error {
	_, _, err := c.run(
		"apply",
		arg("input", "false"),
		arg("var", fmt.Sprintf("lambda_name=%s", function.Name)),
		arg("var", fmt.Sprintf("lambda_file=%s", function.ZipFile)),
		arg("var", fmt.Sprintf("environment=%s", function.Environment)),
		arg("var", fmt.Sprintf("bundle_name=%s", function.BundleName)),
		arg("var", fmt.Sprintf("handler=%s", function.Handler)),
		arg("var", fmt.Sprintf("runtime=%s", function.Runtime)),
		arg("var", fmt.Sprintf("private_subnet_ids=%s", os.Getenv("PrivateSubnetIDs"))),
		arg("auto-approve"),
	)
	if err != nil {
		return errors.Wrap(err, "failed to invoke terraform apply")
	}

	return nil
}

// ApplyTarget invokes terraform apply with the given target.
func (c *Cmd) ApplyTarget(target string) error {
	_, _, err := c.run(
		"apply",
		arg("input", "false"),
		arg("target", target),
		arg("auto-approve"),
	)
	if err != nil {
		return errors.Wrap(err, "failed to invoke terraform apply")
	}

	return nil
}

// Destroy invokes terraform destroy.
func (c *Cmd) Destroy() error {
	_, _, err := c.run(
		"destroy",
		"-auto-approve",
	)
	if err != nil {
		return errors.Wrap(err, "failed to invoke terraform destroy")
	}

	return nil
}

// Output invokes terraform output and returns the named value, true if it exists, and an empty
// string and false if it does not.
func (c *Cmd) Output(variable string) (string, bool, error) {
	stdout, _, err := c.run(
		"output",
		"-json",
	)
	if err != nil {
		return string(stdout), false, errors.Wrap(err, "failed to invoke terraform output")
	}

	var outputs map[string]terraformOutput
	err = json.Unmarshal(stdout, &outputs)
	if err != nil {
		return string(stdout), false, errors.Wrap(err, "failed to parse terraform output")
	}

	value, ok := outputs[variable]

	return fmt.Sprintf("%s", value.Value), ok, nil
}

// Version invokes terraform version and returns the value.
func (c *Cmd) Version() (string, error) {
	stdout, _, err := c.run("version")
	trimmed := strings.TrimSuffix(string(stdout), "\n")
	if err != nil {
		return trimmed, errors.Wrap(err, "failed to invoke terraform version")
	}

	return trimmed, nil
}
