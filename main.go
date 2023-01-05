package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pkg/errors"
	"go.uber.org/zap/zapcore"

	awsTools "github.com/mattermost/mattermost-apps/internal/tools/aws"
	exechelper "github.com/mattermost/mattermost-apps/internal/tools/exechelper"
	terraform "github.com/mattermost/mattermost-apps/internal/tools/terraform"
	model "github.com/mattermost/mattermost-apps/model"
	apps "github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	appsutils "github.com/mattermost/mattermost-plugin-apps/utils"
)

const (
	// manifestFileName is the name of the bundle's manifest file
	manifestFileName = "manifest.json"
)

func main() {
	logger := appsutils.MustMakeCommandLogger(zapcore.InfoLevel)

	err := checkEnvVariables()
	if err != nil {
		logger.WithError(err).Errorf("Environment variables were not set")
		err = sendMattermostErrorNotification(err, "Mattermost apps deployment failed.")
		if err != nil {
			logger.WithError(err).Errorf("Failed to send Mattermost error notification")
		}
		os.Exit(1)
	}

	session, err := awsTools.GetAssumeRoleSession(os.Getenv("AppsAssumeRole"))
	if err != nil {
		logger.WithError(err).Errorf("Failed to get assumed role session")
		err = sendMattermostErrorNotification(err, "Mattermost apps deployment failed.")
		if err != nil {
			logger.WithError(err).Errorf("Failed to send Mattermost error notification")
		}
		os.Exit(1)
	}

	bundles, err := awsTools.GetBundles(os.Getenv("AppsBundleBucketName"), session, logger)
	if err != nil {
		logger.WithError(err).Errorf("Failed to get app bundles")
		err = sendMattermostErrorNotification(err, "Mattermost apps deployment failed.")
		if err != nil {
			logger.WithError(err).Errorf("Failed to send Mattermost error notification")
		}
		os.Exit(1)
	}

	for _, bundle := range bundles {
		var deployData *apps.DeployData
		deployData, err = handleBundleDeployment(bundle, session, logger)
		if err != nil {
			logger.WithError(err).Errorf("Failed to deploy bundle")
			err = sendMattermostErrorNotification(err, "Mattermost apps deployment failed.")
			if err != nil {
				logger.WithError(err).Errorf("Failed to send Mattermost error notification")
			}
			continue
		}

		err = sendAppDeploymentNotification(deployData, bundle)
		if err != nil {
			logger.WithError(err).Errorf("Failed to send Mattermost error notification")
		}
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

func handleBundleDeployment(bundle string, session *session.Session, logger appsutils.Logger) (*apps.DeployData, error) {
	bundleName := strings.TrimSuffix(bundle, ".zip")

	logger = logger.With("bundle", bundleName)

	logger.Infof("Downloading bundle from s3")
	err := awsTools.DownloadS3Object(os.Getenv("AppsBundleBucketName"), bundle, os.Getenv("TempDir"), session)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get s3 object")
	}

	logger.Infof("Unzipping bundle")
	err = exechelper.UnzipBundle(os.Getenv("TempDir"), bundle)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unzip the bundle")
	}

	logger.Infof("Getting bundle details")
	provisionData, err := apps.GetDeployDataFromFile(path.Join(os.Getenv("TempDir"), bundle), logger)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get bundle details for bundle")
	}

	logger.Infof("Uploading bundle assets in %s", os.Getenv("StaticBucket"))
	err = awsTools.UploadStaticFiles(provisionData.StaticFiles, bundleName, logger)
	if err != nil {
		return provisionData, errors.Wrap(err, "failed to upload bundle assets")
	}

	logger.Infof("Uploading bundle manifest file in %s", os.Getenv("StaticBucket"))
	err = awsTools.UploadManifestFile(provisionData.ManifestKey, manifestFileName, bundleName, logger)
	if err != nil {
		return provisionData, errors.Wrap(err, "failed to upload bundle manifest file")
	}

	logger.Infof("Deploying lambdas")
	err = deployLambdas(logger, provisionData.LambdaFunctions, bundle, bundleName)
	if err != nil {
		return provisionData, errors.Wrap(err, "failed to deploy lambda functions for bundle")
	}

	logger.Infof("Tagging bundle object %s as deployed", bundleName)
	err = awsTools.PutDeployedObjectTag(os.Getenv("AppsBundleBucketName"), bundle, session)
	if err != nil {
		return provisionData, errors.Wrap(err, "failed to tag bundle object as deployed")
	}

	logger.Infof("Removing local files for bundle %s", bundleName)
	err = exechelper.RemoveLocalFiles([]string{path.Join(os.Getenv("TempDir"), bundle), path.Join(os.Getenv("TempDir"), bundleName)}, logger)
	if err != nil {
		return provisionData, errors.Wrap(err, "failed to delete local files")
	}

	return provisionData, nil
}

func deployLambdas(logger utils.Logger, lambdaFunctions map[string]apps.FunctionData, bundle, bundleName string) error {
	for zipFile, lambda := range lambdaFunctions {
		logger := logger.With("lambda_name", lambda.Name)

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
			logger.Infof("applying Terraform template")
			err = tf.Apply(function)
			if err != nil {
				return errors.Wrap(err, "failed to run Terraform apply")
			}
			logger.Infof("Successfully deployed lambda function")
			continue
		}
		err = tf.Plan(function)
		if err != nil {
			return errors.Wrap(err, "failed to run Terraform plan")
		}
		logger.Infof("Successfully ran Terraform plan")

	}

	logger.Infof("Successfully deployed all lambda functions")

	return nil
}
