package storage

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	basePath      string
	uploadBaseURL string
}

func NewLocalStorage(basePath, uploadBaseURL string) *LocalStorage {
	os.MkdirAll(filepath.Join(basePath, "raw"), 0755)
	os.MkdirAll(filepath.Join(basePath, "hls"), 0755)
	return &LocalStorage{basePath: basePath, uploadBaseURL: uploadBaseURL}
}

func (s *LocalStorage) SaveFileToRawStorage(r io.Reader, filename string) error {
	dst := filepath.Join(s.basePath, "raw", filename)
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

func (s *LocalStorage) GetFile(ctx context.Context, key string) (io.ReadCloser, error) {
	slog.Info(filepath.Join(s.basePath, key))
	return os.Open(filepath.Join(s.basePath, key))
}

func (s *LocalStorage) GetRawKey(filename string) string {
	return fmt.Sprintf("raw/%s", filename)
}

func (s *LocalStorage) GetHlsKey(id string) string {
	return fmt.Sprintf("hls/%s", id)
}

func (s *LocalStorage) FileExists(_ context.Context, key string) (bool, error) {
	_, err := os.Stat(filepath.Join(s.basePath, key))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *LocalStorage) GenerateRawUploadUrl(_ context.Context, key string) (string, error) {
	filename := filepath.Base(key)
	return fmt.Sprintf("%s/videos/upload-raw/%s", s.uploadBaseURL, filename), nil
}

func (s *LocalStorage) GetRawFilePath(filename string) string {
	return filepath.Join(s.basePath, "raw", filename)
}

func (s *LocalStorage) GetHlsFilePath(id string) string {
	path := filepath.Join(s.basePath, "hls", id)
	os.MkdirAll(path, 0755)
	return path
}

func (s *LocalStorage) DownloadFile(_ context.Context, _ string, _ string) error {
	return nil
}

func (s *LocalStorage) UploadFile(_ context.Context, _ string, _ string) error {
	return nil
}

func (s *LocalStorage) DeleteRawLocalFile(_ context.Context, localPath string) error {
	return os.Remove(localPath)
}

func (s *LocalStorage) DeleteHlsLocalFolder(_ string) error {
	return nil
}

func (s *LocalStorage) GeneratePresignedGetURL(_ context.Context, key string) (string, error) {
	return fmt.Sprintf("%s/videos/%s", s.uploadBaseURL, key), nil
}
