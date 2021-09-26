//Package simplestorageservice simple storage service functions S3
package simplestorageservice

import (
	"fmt"

	"goblog/internal/config"

	"github.com/aws/aws-sdk-go/aws"
	awsSession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// DeleteS3Object Deletes and object from s3
func DeleteS3Object(key string) {

	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Printf("could not get configuration object %s", (err))
		return
	}

	// Create a single AWS session (we can re use this if we're uploading many files)
	s, err := awsSession.NewSession(&aws.Config{Region: aws.String("us-west-2")})
	if err != nil {
		fmt.Printf("Error creating session to s3 with error %s\n", err)
		return
	}

	svc := s3.New(s)

	_, err = svc.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(cfg.S3Bucket), Key: aws.String(key)})
	if err != nil {
		fmt.Printf("Unable to delete object %q from bucket %q, %v", key, cfg.S3Bucket, err)
		return
	}

	err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(cfg.S3Bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		fmt.Printf("Unable to wait on delete of object %q from bucket %q, %v", key, cfg.S3Bucket, err)
		return
	}

	return
}
