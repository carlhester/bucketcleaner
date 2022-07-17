package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type s3APIClient interface {
	ListObjects(input *s3.ListObjectsInput) (*s3.ListObjectsOutput, error)
	ListObjectVersions(input *s3.ListObjectVersionsInput) (*s3.ListObjectVersionsOutput, error)
	DeleteObject(input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)
	DeleteBucket(input *s3.DeleteBucketInput) (*s3.DeleteBucketOutput, error)
	PutBucketPolicy(input *s3.PutBucketPolicyInput) (*s3.PutBucketPolicyOutput, error)
	WaitUntilBucketNotExists(input *s3.HeadBucketInput) error
}

type app struct {
	s3Client   s3APIClient
	bucketName string
}

func denyAllBucketPolicy(bucketName string) string {
	policy := `{
		"Version": "2012-10-17",
		"Id": "S3DenyAllPutObjectPolicy",
		"Statement": [
			{
				"Effect": "Deny",
				"Principal": "*",
				"Action": "s3:PutObject",
				"Resource": [
					"arn:aws:s3:::%[1]s",
					"arn:aws:s3:::%[1]s/*"
				]
			}
		]
	}`

	return fmt.Sprintf(policy, bucketName)

}

func (a *app) run() error {
	// apply a deny all bucket policy to prevent further writes
	_, err := a.s3Client.PutBucketPolicy(&s3.PutBucketPolicyInput{
		Bucket: aws.String(a.bucketName),
		Policy: aws.String(denyAllBucketPolicy(a.bucketName)),
	})

	if err != nil {
		return err
	}

	// Get the list of objects in the bucket
	listObjectOutput, err := a.s3Client.ListObjects(
		&s3.ListObjectsInput{
			Bucket: aws.String(a.bucketName),
		})

	if err != nil {
		return err
	}

	for _, object := range listObjectOutput.Contents {
		input := &s3.ListObjectVersionsInput{
			Bucket: aws.String(a.bucketName),
			Prefix: object.Key,
		}

		result, err := a.s3Client.ListObjectVersions(input)
		if err != nil {
			return err
		}

		for _, version := range result.Versions {
			_, err := a.s3Client.DeleteObject(&s3.DeleteObjectInput{
				Bucket:    aws.String(a.bucketName),
				Key:       version.Key,
				VersionId: version.VersionId,
			})
			if err != nil {
				return err
			}
		}

	}
	_, err = a.s3Client.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(a.bucketName),
	})
	if err != nil {
		return err
	}

	err = a.s3Client.WaitUntilBucketNotExists(&s3.HeadBucketInput{
		Bucket: aws.String(a.bucketName),
	})
	if err != nil {
		return err
	}

	return nil

}

func main() {

	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s BUCKETNAME REGION <really>\n\n", os.Args[0])
		return
	}

	bucket := os.Args[1]
	region := os.Args[2]

	if !(len(os.Args) == 4 && os.Args[3] == "really") {
		fmt.Printf("Confirm you want to delete bucket \"%s\"?\nType yes to continue: ", bucket)
		input := bufio.NewScanner(os.Stdin)
		input.Scan()
		if input.Text() != "yes" {
			fmt.Println("aborting")
			return
		}
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region:              aws.String(region),
		STSRegionalEndpoint: endpoints.RegionalSTSEndpoint,
	}))

	s3Svc := s3.New(sess)

	app := &app{
		s3Client:   s3Svc,
		bucketName: bucket,
	}

	if err := app.run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)

}
