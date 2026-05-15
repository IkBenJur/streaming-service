package videoProcessing

import (
	"mime/multipart"
	"os"
	"path/filepath"

	repo "github.com/IkBenJur/streaming-service/internal/postgres/sqlc"
	"github.com/gin-gonic/gin"
)

type VideoProcessingService interface {
	repo.Querier
	SaveFileToRawStorage(c *gin.Context, file *multipart.FileHeader, filename string) error
	GetChunk(name string, start, end int64) ([]byte, error)
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

func (s *LocalStorage) SaveFileToRawStorage(c *gin.Context, file *multipart.FileHeader, filename string) error {
	dst := filepath.Join(s.basePath, "raw", filename)
	err := c.SaveUploadedFile(file, dst)
	return err
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
