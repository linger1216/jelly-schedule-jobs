package s3

import (
	"context"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	awsSession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	Region    = "cn-east-2"
	AccessKey = "13EEE71557A91B8C4561D7282A75A647"
	SecretKey = "58358E170934778DD2E2DA493DBD07F6"
	Bucket    = "wz-location"
	EndPoint  = ""
)

type S3Config struct {
	Region    string `json:"region,omitempty" yaml:"region" `
	AccessKey string `json:"accessKey,omitempty" yaml:"accessKey" `
	SecretKey string `json:"secretKey,omitempty" yaml:"secretKey" `
	Bucket    string `json:"bucket,omitempty" yaml:"bucket" `
	EndPoint  string `json:"endPoint,omitempty" yaml:"endPoint" `
}

type S3Svc struct {
	svc     *s3.S3
	session *awsSession.Session
	bucket  string
}

func NewS3Svc(conf S3Config) *S3Svc {
	return newS3svc(conf.Region, conf.Bucket, conf.EndPoint, conf.AccessKey, conf.SecretKey)
}

func newS3svc(region, bucket, endpoint, accessKey, secretKey string) *S3Svc {
	sess, err := awsSession.NewSession(&aws.Config{
		Region:      aws.String(region),
		Endpoint:    aws.String(endpoint),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		panic(err)
	}

	return &S3Svc{
		svc:     s3.New(sess),
		session: sess,
		bucket:  bucket,
	}
}

func (s3svc *S3Svc) ListObjects(ctx context.Context, prefix string) ([]string, error) {
	var objkeys []string

	output, err := s3svc.svc.ListObjectsWithContext(ctx, &s3.ListObjectsInput{
		Bucket: aws.String(s3svc.bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, err
	}

	for _, content := range output.Contents {
		objkeys = append(objkeys, aws.StringValue(content.Key))
	}
	return objkeys, nil
}

func (s3svc *S3Svc) DownloadObject(key, localName string) (int64, error) {
	downloader := s3manager.NewDownloader(s3svc.session)
	file, err := os.Create(localName)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	nBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(s3svc.bucket),
			Key:    aws.String(key),
		})
	if err != nil {
		return 0, err
	}
	return nBytes, nil
}

func (s3svc *S3Svc) UploadObject(key string, body io.Reader) error {
	uploader := s3manager.NewUploader(s3svc.session, func(u *s3manager.Uploader) { u.Concurrency = 1 })
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s3svc.bucket),
		Key:    aws.String(key),
		Body:   body,
	})
	return err
}

func (s3svc *S3Svc) DeleteObject(key string) error {
	_, err := s3svc.svc.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(s3svc.bucket), Key: aws.String(key)})
	if err != nil {
		return err
	}

	return s3svc.svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(s3svc.bucket),
		Key:    aws.String(key),
	})
}
