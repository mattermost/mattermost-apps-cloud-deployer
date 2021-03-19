package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pkg/errors"

	awsTools "github.com/mattermost/mattermost-apps/internal/tools/aws"
	exechelper "github.com/mattermost/mattermost-apps/internal/tools/exechelper"
	terraform "github.com/mattermost/mattermost-apps/internal/tools/terraform"
	model "github.com/mattermost/mattermost-apps/model"
	apps "github.com/mattermost/mattermost-plugin-apps/aws"
)

func main() {
	err := checkEnvVariables()
	if err != nil {
		logger.WithError(err).Error("Environment variables were not set")
		err = sendMattermostErrorNotification(err, "The Mattermost apps deployment failed.")
		if err != nil {
			logger.WithError(err).Error("Failed to send Mattermost error notification")
		}
		os.Exit(1)
	}

	session, err := awsTools.GetAssumeRoleSession(os.Getenv("AppsAssumeRole"))
	if err != nil {
		logger.WithError(err).Error("Failed to get assumed role session")
		err = sendMattermostErrorNotification(err, "The Mattermost apps deployment failed.")
		if err != nil {
			logger.WithError(err).Error("Failed to send Mattermost error notification")
		}
		os.Exit(1)
	}

	bundles, err := awsTools.GetBundles(os.Getenv("AppsBundleBucketName"), session)
	if err != nil {
		logger.WithError(err).Error("Failed to get app bundles")
		err = sendMattermostErrorNotification(err, "The Mattermost apps deployment failed.")
		if err != nil {
			logger.WithError(err).Error("Failed to send Mattermost error notification")
		}
		os.Exit(1)
	}

	var deployedBundles []string
	for _, bundle := range bundles {
		err = handleBundleDeployment(bundle, session)
		if err != nil {
			logger.WithError(err).Error("Failed to deploy bundle")
			err = sendMattermostErrorNotification(err, "The Mattermost apps deployment failed.")
			if err != nil {
				logger.WithError(err).Error("Failed to send Mattermost error notification")
			}
			return
		}
		deployedBundles = append(deployedBundles, strings.TrimSuffix(bundle, ".zip"))
	}

	err = sendMattermostNotification(deployedBundles, "The Mattermost apps were successfully deployed/updated")
	if err != nil {
		logger.WithError(err).Error("Failed to send Mattermost error notification")
	}
}

func checkEnvVariables() error {
	var envVariables = []string{
		"AppsBundleBucketName",
		"TempDir",
		"TerraformTemplateDir",
		"TerraformStateBucket",
		"AppsAssumeRole",
		"StaticBucket",
		"Environment",
		"TerraformApply",
		"MattermostNotificationsHook",
		"MattermostAlertsHook",
		"PrivateSubnetIDs",
	}

	for _, envVar := range envVariables {
		if os.Getenv(envVar) == "" {
			return errors.Errorf("Environment variable %s was not set", envVar)
		}
	}

	return nil
}

func handleBundleDeployment(bundle string, session *session.Session) error {
	bundleName := strings.TrimSuffix(bundle, ".zip")

	logger.Infof("Downloading bundle %s from s3", bundleName)
	err := awsTools.DownloadS3Object(os.Getenv("AppsBundleBucketName"), bundle, os.Getenv("TempDir"), session)
	if err != nil {
		return errors.Wrap(err, "failed to get s3 object")
	}

	logger.Infof("Unzipping bundle %s", bundleName)
	err = exechelper.UnzipBundle(os.Getenv("TempDir"), bundle)
	if err != nil {
		return errors.Wrap(err, "failed to unzip the bundle")
	}

	logger.Infof("Getting bundle details from bundle %s", bundleName)
	provisionData, err := apps.GetProvisionDataFromFile(fmt.Sprintf("%s/%s", os.Getenv("TempDir"), bundle), nil)
	if err != nil {
		return errors.Wrap(err, "failed to get bundle details for bundle")
	}

	logger.Infof("Uploading bundle %s assets in %s", bundleName, os.Getenv("StaticBucket"))
	err = awsTools.UploadStaticFiles(provisionData.StaticFiles, bundleName, logger)
	if err != nil {
		return errors.Wrap(err, "failed to upload bundle assets")
	}

	logger.Infof("Deploying lambdas from bundle %s", bundleName)
	err = deployLambdas(provisionData.LambdaFunctions, bundle, bundleName)
	if err != nil {
		return errors.Wrap(err, "failed to get deploy lambda functions for bundle")
	}

	logger.Infof("Removing local files for bundle %s", bundleName)
	err = exechelper.RemoveLocalFiles([]string{fmt.Sprintf("%s/%s", os.Getenv("TempDir"), bundle), fmt.Sprintf("%s/%s", os.Getenv("TempDir"), bundleName)}, logger)
	if err != nil {
		return errors.Wrap(err, "failed to delete local files")
	}
	return nil
}

func deployLambdas(lambdaFunctions map[string]apps.FunctionData, bundle, bundleName string) error {
	for zipFile, lambda := range lambdaFunctions {
		function := model.Function{
			Name:        lambda.Name,
			Environment: os.Getenv("Environment"),
			Runtime:     lambda.Runtime,
			Handler:     lambda.Handler,
			ZipFile:     fmt.Sprintf("%s.zip", zipFile),
			BundleName:  bundleName,
		}

		tf, err := terraform.New(os.Getenv("TerraformTemplateDir"), os.Getenv("TerraformStateBucket"), logger)
		if err != nil {
			return errors.Wrap(err, "failed to initiate Terraform")
		}

		err = tf.Init(lambda.Name)
		if err != nil {
			return errors.Wrap(err, "failed to run Terraform init")
		}

		if os.Getenv("TerraformApply") == "true" {
			logger.Info("applying Terraform template")
			err = tf.Apply(function)
			if err != nil {
				return errors.Wrap(err, "failed to run Terraform apply")
			}
			logger.Infof("Successfully deployed lambda function %s", function.Name)
			continue
		}
		err = tf.Plan(function)
		if err != nil {
			return errors.Wrap(err, "failed to run Terraform plan")
		}
		logger.Info("Successfully ran Terraform plan")

	}
	logger.Infof("Successfully deployed all lambda functions for bundle %s", bundleName)
	return nil
}
