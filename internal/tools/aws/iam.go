package aws

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

// GetAssumeRoleSession assumes an IAM role and returns the session.
func GetAssumeRoleSession(iamRole string) (*session.Session, error) {
	svcSTS := sts.New(session.New())

	input := &sts.AssumeRoleInput{
		RoleArn:         aws.String(iamRole),
		RoleSessionName: aws.String("AssumeRoleSession"),
	}

	assumeRole, err := svcSTS.AssumeRole(input)
	if err != nil {
		return nil, err
	}

	provider := NewAssumeRoleCredentialsProvider(assumeRole.Credentials)
	session, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewCredentials(provider),
	})
	if err != nil {
		return nil, err
	}
	return session, nil

}

// NewAssumeRoleCredentialsProvider returns AssumeRoleCredentialsProvider using provided credentials.
func NewAssumeRoleCredentialsProvider(credentials *sts.Credentials) *AssumeRoleCredentialsProvider {
	return &AssumeRoleCredentialsProvider{
		AssumeRoleCredentials: credentials,
	}
}

// AssumeRoleCredentialsProvider describes assume role credentials.
type AssumeRoleCredentialsProvider struct {
	AssumeRoleCredentials *sts.Credentials
}

// IsExpired checks if the assume role session has expired.
func (c AssumeRoleCredentialsProvider) IsExpired() bool {
	return c.AssumeRoleCredentials.Expiration.After(time.Now())
}

// Retrieve returns the creds values.
func (c AssumeRoleCredentialsProvider) Retrieve() (credentials.Value, error) {
	return credentials.Value{
		AccessKeyID:     *c.AssumeRoleCredentials.AccessKeyId,
		SecretAccessKey: *c.AssumeRoleCredentials.SecretAccessKey,
		SessionToken:    *c.AssumeRoleCredentials.SessionToken,
		ProviderName:    "AssumeRoleCredentialsProvider",
	}, nil
}
