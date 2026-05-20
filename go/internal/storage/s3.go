package storage

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Storage struct {
	client        s3.Client
	presignClient s3.PresignClient
}

func NewS3Storage(config aws.Config) *S3Storage {
	client := s3.NewFromConfig(config)

	return &S3Storage{
		client:        *client,
		presignClient: *s3.NewPresignClient(client),
	}
}
