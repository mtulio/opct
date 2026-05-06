package baseline

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	awsMaxRetries    = 10
	awsRetryMinDelay = 1 * time.Second
	awsRetryMaxDelay = 30 * time.Second
)

// newAWSSession creates a shared AWS session with retry and backoff configuration.
// SDK v1 does not honor AWS_MAX_ATTEMPTS/AWS_RETRY_MODE env vars (those are SDK v2),
// so retry behavior must be configured explicitly in code.
func newAWSSession(region string) (*session.Session, error) {
	return session.NewSession(&aws.Config{
		Region:     aws.String(region),
		MaxRetries: aws.Int(awsMaxRetries),
		Retryer: client.DefaultRetryer{
			NumMaxRetries:    awsMaxRetries,
			MinRetryDelay:    awsRetryMinDelay,
			MaxRetryDelay:    awsRetryMaxDelay,
			MinThrottleDelay: awsRetryMinDelay,
			MaxThrottleDelay: awsRetryMaxDelay,
		},
	})
}

// createS3Client creates an S3 client with the specified region.
func createS3Client(region string) (*s3.S3, *s3manager.Uploader, error) {
	sess, err := newAWSSession(region)
	if err != nil {
		return nil, nil, err
	}

	svc := s3.New(sess)
	uploader := s3manager.NewUploader(sess)

	return svc, uploader, nil
}

// createCloudFrontClient creates a CloudFront client with the specified region.
func createCloudFrontClient(region string) (*cloudfront.CloudFront, error) {
	sess, err := newAWSSession(region)
	if err != nil {
		return nil, err
	}

	svc := cloudfront.New(sess)
	return svc, nil
}

// checkBucketExists checks if the bucket exists in the S3 storage.
func checkBucketExists(svc *s3.S3, bucket string) (bool, error) {
	_, err := svc.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return false, fmt.Errorf("failed to check if bucket exists: %v", err)
	}
	return true, nil
}
