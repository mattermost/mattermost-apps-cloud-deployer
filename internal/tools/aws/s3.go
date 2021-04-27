package aws

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	apps "github.com/mattermost/mattermost-plugin-apps/aws"
	log "github.com/sirupsen/logrus"
)

// DownloadS3Object downloads the specified object from the specified S3 bucket.
func DownloadS3Object(bucketName, object, dir string, session *session.Session) error {
	file, err := os.Create(path.Join(dir, object))
	if err != nil {
		return err
	}

	defer file.Close()

	downloader := s3manager.NewDownloader(session)

	_, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(object),
		})
	if err != nil {
		return err
	}

	return nil
}

// UploadStaticFiles is used to upload static files to the static S3 bucket.
func UploadStaticFiles(staticFiles map[string]apps.AssetData, bundleName string, logger log.FieldLogger) error {
	for staticFile, staticKey := range staticFiles {
		bundleDir := path.Join(os.Getenv("TempDir"), bundleName)
		fileDir := path.Join(bundleDir, "static", staticFile)
		file, err := os.Open(fileDir)
		if err != nil {
			return err
		}

		defer file.Close()

		uploader := s3manager.NewUploader(session.New())
		_, err = uploader.Upload(
			&s3manager.UploadInput{
				Bucket: aws.String(os.Getenv("StaticBucket")),
				Key:    aws.String(staticKey.Key),
				Body:   file,
			})
		if err != nil {
			return err
		}

		logger.Infof("Uploaded file %s with object name %s", file.Name(), staticKey.Key)
	}
	return nil
}

// UploadManifestFile is used to upload the manifest file to the static S3 bucket.
func UploadManifestFile(manifestKey, manifestFileName, bundleName string, logger log.FieldLogger) error {
	bundleDir := path.Join(os.Getenv("TempDir"), bundleName)
	fileDir := path.Join(bundleDir, manifestFileName)
	file, err := os.Open(fileDir)
	if err != nil {
		return err
	}

	defer file.Close()

	uploader := s3manager.NewUploader(session.New())
	_, err = uploader.Upload(
		&s3manager.UploadInput{
			Bucket: aws.String(os.Getenv("StaticBucket")),
			Key:    aws.String(manifestKey),
			Body:   file,
		})
	if err != nil {
		return err
	}

	logger.Infof("Uploaded file %s with object name %s", file.Name(), manifestKey)
	return nil
}

// GetBundles is used to get all app bundles from a S3 bucket.
// TODO: Limit of 1000 objects per API call should be handled.
func GetBundles(bucketName string, session *session.Session, logger log.FieldLogger) ([]string, error) {
	var bundles []string

	svc := s3.New(session)
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	}

	result, err := svc.ListObjectsV2(input)
	if err != nil {
		return nil, err
	}

	for _, content := range result.Contents {
		if strings.HasSuffix(*content.Key, ".zip") {
			isDeployed, err := IsBundleDeployed(bucketName, *content.Key, session)
			if err != nil {
				return nil, err
			}
			if !isDeployed {
				bundles = append(bundles, *content.Key)
			} else {
				logger.Infof("Bundle %s is already deployed", *content.Key)
			}
		}
	}
	return bundles, nil
}

// IsBundleDeployed checks the bundle object tag to check if it got deployed before.
func IsBundleDeployed(bucketName, objectKey string, session *session.Session) (bool, error) {
	svc := s3.New(session)
	result, err := svc.GetObjectTagging(&s3.GetObjectTaggingInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return false, err
	}

	for _, tag := range result.TagSet {
		if *tag.Key == fmt.Sprintf("deployed_%s", os.Getenv("Environment")) && *tag.Value == "true" {
			return true, nil
		}
	}

	return false, nil
}

// PutDeployedObjectTag adds a tag to specify that the bundle was deployed.
func PutDeployedObjectTag(bucketName, objectKey string, session *session.Session) error {
	svc := s3.New(session)

	input := &s3.PutObjectTaggingInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Tagging: &s3.Tagging{
			TagSet: []*s3.Tag{
				{
					Key:   aws.String(fmt.Sprintf("deployed_%s", os.Getenv("Environment"))),
					Value: aws.String("true"),
				},
			},
		},
	}

	_, err := svc.PutObjectTagging(input)
	if err != nil {
		return err
	}
	return nil
}
