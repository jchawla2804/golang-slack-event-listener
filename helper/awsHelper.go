package helper

import (
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func GetAwsSession() *session.Session {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY"), os.Getenv("AWS_SECRET_KEY"), ""),
			Region:      aws.String("us-west-2"),
		},
	})

	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	return sess
}

func GetAWSKey(controlNumber, bucketName, prefix string) {
	sess := GetAwsSession()
	s3svc := s3.New(sess)
	listInputHl7 := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(prefix + controlNumber),
	}

	listObjectOutput, err := s3svc.ListObjectsV2(listInputHl7)
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
	if *listObjectOutput.KeyCount == 0 {
		log.Println("Key Does not exist for given prefix " + prefix + controlNumber)
		os.Exit(1)
	}

	err = DownloadObject(*listObjectOutput.Contents[0].Key, bucketName, sess)
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
}

func DownloadObject(keyName, bucketName string, sess *session.Session) error {
	fileName := strings.Split(keyName, "/")[3]
	file, err := os.Create(fileName)
	if err != nil {
		log.Println("Unable to create a new file")
		os.Exit(1)
	}

	defer file.Close()

	s3ObjectInput := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(keyName),
	}

	downloader := s3manager.NewDownloader(sess)
	_, err = downloader.Download(file, s3ObjectInput)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	return nil

}
