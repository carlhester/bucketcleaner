package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type s3APIClient interface {
	ListObjects(input *s3.ListObjectsInput) (*s3.ListObjectsOutput, error)
	ListObjectsPages(input *s3.ListObjectsInput, fn func(p *s3.ListObjectsOutput, lastPage bool) bool) error
	ListObjectVersions(input *s3.ListObjectVersionsInput) (*s3.ListObjectVersionsOutput, error)
	ListObjectVersionsPages(input *s3.ListObjectVersionsInput, fn func(p *s3.ListObjectVersionsOutput, lastPage bool) bool) error
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

	// List all objects in the bucket
	bucketContents := []*s3.Object{}
	err = a.s3Client.ListObjectsPages(&s3.ListObjectsInput{
		Bucket: aws.String(a.bucketName),
	},
		func(page *s3.ListObjectsOutput, lastPage bool) bool {
			bucketContents = append(bucketContents, page.Contents...)
			return !lastPage
		})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			// If the bucket doesn't exist, consider it a success and return
			if aerr.Code() == s3.ErrCodeNoSuchBucket {
				return nil
			}
		}
		return fmt.Errorf("failed to list objects in bucket %s: err: %v", a.bucketName, err)
	}

	// Get every version of each object
	objectVersions := []*s3.ObjectVersion{}
	for _, object := range bucketContents {
		err = a.s3Client.ListObjectVersionsPages(&s3.ListObjectVersionsInput{
			Bucket: aws.String(a.bucketName),
			Prefix: object.Key,
		},
			func(page *s3.ListObjectVersionsOutput, lastPage bool) bool {
				objectVersions = append(objectVersions, page.Versions...)
				return !lastPage
			})

		if err != nil {
			return fmt.Errorf("failed to list object versions: %v. err: %v", object, err)
		}
	}

	// for every object version, delete it
	for _, version := range objectVersions {
		_, err := a.s3Client.DeleteObject(&s3.DeleteObjectInput{
			Bucket:    aws.String(a.bucketName),
			Key:       version.Key,
			VersionId: version.VersionId,
		})
		if err != nil {
			return fmt.Errorf("failed to delete object: %v. err: %v", version, err)
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
