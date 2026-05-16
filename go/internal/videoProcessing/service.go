package videoProcessing

import (
	"io"
	"os"
	"path/filepath"

	repo "github.com/IkBenJur/streaming-service/internal/postgres/sqlc"
)

type VideoProcessingService interface {
	repo.Querier
	SaveFileToRawStorage(r io.Reader, filename string) error
	GetChunk(name string, start, end int64) ([]byte, error)
	RawFilePath(filename string) string
	HLSOutputPath(id string) (string, error)
}

type LocalStorage struct {
	repo.Querier
	basePath string
}

func NewLocalStorage(basePath string, queries repo.Querier) *LocalStorage {
	return &LocalStorage{
		Querier:  queries,
		basePath: basePath,
	}
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

func (s *LocalStorage) RawFilePath(filename string) string {
	return filepath.Join(s.basePath, "raw", filename)
}

func (s *LocalStorage) HLSOutputPath(id string) (string, error) {
	path := filepath.Join(s.basePath, "hls", id)
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", err
	}
	return path, nil
}

func (s *LocalStorage) GetChunk(name string, start, end int64) ([]byte, error) {
	file, err := os.Open(filepath.Join(s.basePath, filepath.Base(name)))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf := make([]byte, end-start+1)
	_, err = file.ReadAt(buf, start)
	return buf, err
}
