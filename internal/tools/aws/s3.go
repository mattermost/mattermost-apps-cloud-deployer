package aws

import (
	"fmt"
	"os"
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
	file, err := os.Create(fmt.Sprintf("%s/%s", dir, object))
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
		bundleDir := fmt.Sprintf("%s/%s", os.Getenv("TempDir"), bundleName)
		fileDir := fmt.Sprintf("%s/static/%s", bundleDir, staticFile)
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

// GetBundles is used to get all app bundles from a S3 bucket.
func GetBundles(bucketName string, session *session.Session) ([]string, error) {
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
			bundles = append(bundles, *content.Key)
		}
	}
	return bundles, nil
}
