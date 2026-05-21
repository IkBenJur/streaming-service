package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Storage struct {
	client        s3.Client
	presignClient s3.PresignClient
	bucketName    string
}

func NewS3Storage(config aws.Config,
	bucketName string) *S3Storage {
	client := s3.NewFromConfig(config)
	return &S3Storage{
		client:        *client,
		presignClient: *s3.NewPresignClient(client),
		bucketName:    bucketName,
	}
}

func GetRawKey(filename string) string {
	return fmt.Sprintf("raw/%s", filename)
}

func (s3Storage *S3Storage) GenerateRawUploadUrl(ctx context.Context, key string) (string, error) {
	req, err := s3Storage.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s3Storage.bucketName),
		Key:    aws.String(key),
	}, func(po *s3.PresignOptions) {
		po.Expires = 1 * time.Hour
	})

	return req.URL, err
}

func (s3Storage *S3Storage) FileExists(ctx context.Context, key string) (bool, error) {
	_, err := s3Storage.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s3Storage.bucketName),
		Key:    aws.String(key),
	})

	var notFound *types.NotFound
	if err != nil && errors.As(err, &notFound) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}
