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

func main() {

	if len(os.Args) != 3 {
		fmt.Println("Usage: go run main.go <bucket> <region>")
	}

	bucket := os.Args[1]
	region := os.Args[2]

	fmt.Printf("Are you sure you want to delete bucket %s? Type yes to continue: ", bucket)
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	if input.Text() != "yes" {
		fmt.Println("aborting")
		os.Exit(0)
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region:              aws.String(region),
		STSRegionalEndpoint: endpoints.RegionalSTSEndpoint,
	}))

	svc := s3.New(sess)

	listObjectOutput, err := svc.ListObjects(
		&s3.ListObjectsInput{
			Bucket: aws.String(bucket),
		})

	if err != nil {
		panic(err)
	}

	for _, object := range listObjectOutput.Contents {
		input := &s3.ListObjectVersionsInput{
			Bucket: aws.String(bucket),
			Prefix: object.Key,
		}

		result, err := svc.ListObjectVersions(input)
		if err != nil {
			panic(err)
		}

		for _, version := range result.Versions {
			_, err := svc.DeleteObject(&s3.DeleteObjectInput{
				Bucket:    aws.String(bucket),
				Key:       version.Key,
				VersionId: version.VersionId,
			})
			if err != nil {
				panic(err)
			}
			// fmt.Println("Deleted: ", aws.StringValue(version.Key), aws.StringValue(version.VersionId))
			fmt.Print(".")
		}

	}
	_, err = svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		panic(err)
	}

	err = svc.WaitUntilBucketNotExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Deleted bucket: %s\n", bucket)

}
