package storage

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) *LocalStorage {
	os.MkdirAll(filepath.Join(basePath, "raw"), 0755)
	os.MkdirAll(filepath.Join(basePath, "hls"), 0755)
	return &LocalStorage{basePath: basePath}
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
	return filepath.Join(s.basePath, "raw", filename)
}

func (s *LocalStorage) GetHlsKey(id string) string {
	return filepath.Join(s.basePath, "hls", id)
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

func (s *LocalStorage) DeleteHlsLocalFolder(filePath string) error {
	return os.RemoveAll(filePath)
}
