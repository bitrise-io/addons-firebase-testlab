package s3client

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
)

// AWSInterface ...
type AWSInterface interface {
	GeneratePresignedGETURL(key string, expiresIn time.Duration) (string, error)
	GeneratePresignedPUTURL(key string, expiresIn time.Duration, fileSize int64) (string, error)
	GetObjectFromAWS(key string) (string, error)
	GetAWSConfig() Config
}

// Config ...
type Config struct {
	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSBucket          string
}

// AWSProvider ...
type AWSProvider struct {
	Config Config
}

// GetAWSConfig ...
func (p *AWSProvider) GetAWSConfig() Config {
	return p.Config
}

func (p *AWSProvider) createS3Client() (svc *s3.S3, err error) {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(
			p.Config.AWSAccessKeyID,
			p.Config.AWSSecretAccessKey,
			""),
		Region: aws.String(p.Config.AWSRegion),
	})
	if err != nil {
		return nil, errors.Wrap(err, "Session creation failed")
	}

	svc = s3.New(sess)
	return
}

// GeneratePresignedGETURL ...
func (p *AWSProvider) GeneratePresignedGETURL(key string, expiresIn time.Duration) (string, error) {
	svc, err := p.createS3Client()
	if err != nil {
		return "", errors.WithStack(err)
	}

	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(p.Config.AWSBucket),
		Key:    aws.String(key),
	})
	presignedURL, err := req.Presign(expiresIn)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return presignedURL, nil
}

// GeneratePresignedPUTURL ...
func (p *AWSProvider) GeneratePresignedPUTURL(key string, expiresIn time.Duration, fileSize int64) (string, error) {
	svc, err := p.createS3Client()
	if err != nil {
		return "", errors.WithStack(err)
	}

	req, _ := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket:        aws.String(p.Config.AWSBucket),
		Key:           aws.String(key),
		ContentLength: aws.Int64(fileSize),
	})
	presignedURL, err := req.Presign(expiresIn)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return presignedURL, nil
}
