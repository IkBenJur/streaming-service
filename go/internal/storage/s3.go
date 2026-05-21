package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Storage struct {
	client        s3.Client
	presignClient s3.PresignClient
	bucketName    string
	localTempPath string
}

func NewS3Storage(config aws.Config, bucketName string, localTempPath string) *S3Storage {
	client := s3.NewFromConfig(config)
	os.MkdirAll(filepath.Join(localTempPath, "raw"), 0755)
	os.MkdirAll(filepath.Join(localTempPath, "hls"), 0755)
	return &S3Storage{
		client:        *client,
		presignClient: *s3.NewPresignClient(client),
		bucketName:    bucketName,
		localTempPath: localTempPath,
	}
}

func GetRawKey(filename string) string {
	return fmt.Sprintf("raw/%s", filename)
}

func GetHLSKey(id, filename string) string {
	return fmt.Sprintf("hls/%s/%s", id, filename)
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

func (s3Storage *S3Storage) DownloadFile(ctx context.Context, key string, destPath string) error {
	result, err := s3Storage.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s3Storage.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("get object %s: %w", key, err)
	}
	defer result.Body.Close()

	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create file %s: %w", destPath, err)
	}
	defer file.Close()

	if _, err = io.Copy(file, result.Body); err != nil {
		return fmt.Errorf("write file %s: %w", destPath, err)
	}

	return nil
}

func (s3Storage *S3Storage) UploadFile(ctx context.Context, localPath string, key string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open file %s: %w", localPath, err)
	}
	defer file.Close()

	contentType := mime.TypeByExtension(filepath.Ext(localPath))

	_, err = s3Storage.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s3Storage.bucketName),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("upload file %s: %w", key, err)
	}

	return nil
}

func (s3Storage *S3Storage) GetRawKey(filename string) string {
	return fmt.Sprintf("raw/%s", filename)
}

func (s3Storage *S3Storage) GetHlsKey(id string) string {
	return fmt.Sprintf("hls/%s", id)
}

func (s3Storage *S3Storage) GetRawFilePath(filename string) string {
	return filepath.Join(s3Storage.localTempPath, "raw", filename)
}

func (s3Storage *S3Storage) GetHlsFilePath(id string) string {
	path := filepath.Join(s3Storage.localTempPath, "hls", id)
	os.MkdirAll(path, 0755)
	return path
}

func (s3Storage *S3Storage) DeleteRawLocalFile(_ context.Context, localPath string) error {
	return os.Remove(localPath)
}

func (s3Storage *S3Storage) DeleteHlsLocalFolder(filePath string) error {
	return os.RemoveAll(filePath)
}
