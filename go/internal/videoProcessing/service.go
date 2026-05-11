package videoProcessing

import (
	"os"
	"path/filepath"
)

type VideoProcessingService interface {
	GetChunk(name string, start, end int64) ([]byte, error)
}

type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{
		basePath: basePath,
	}
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
