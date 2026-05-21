package videoProcessing

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	repo "github.com/IkBenJur/streaming-service/internal/postgres/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockService struct {
	repo.Querier
	findVideoById func(ctx context.Context, id pgtype.UUID) (repo.Video, error)
	getFile       func(ctx context.Context, key string) (io.ReadCloser, error)
}

func (m *mockService) FindVideoById(ctx context.Context, id pgtype.UUID) (repo.Video, error) {
	return m.findVideoById(ctx, id)
}

func (m *mockService) SaveFileToRawStorage(r io.Reader, filename string) error {
	return nil
}

func (m *mockService) GetFile(ctx context.Context, key string) (io.ReadCloser, error) {
	return m.getFile(ctx, key)
}

type mockTranscoder struct{}

func (m *mockTranscoder) Submit(id pgtype.UUID) {}

func TestGetVideoStream_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ms := &mockService{}
	h := NewHandler(ms, ms, &mockTranscoder{})

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.GET("/videos/:id/stream/:file", h.GetVideoStream)

	c.Request, _ = http.NewRequest("GET", "/videos/not-a-uuid/stream/playlist.m3u8", nil)
	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d; want 400", w.Code)
	}
}

func TestGetVideoStream_StatusNotFinished(t *testing.T) {
	gin.SetMode(gin.TestMode)

	id := uuid.New()
	ms := &mockService{
		findVideoById: func(ctx context.Context, id pgtype.UUID) (repo.Video, error) {
			return repo.Video{
				ID:     id,
				Status: pgtype.UUID{Bytes: [16]byte{1}, Valid: true}, // any non-zero value != uninitialized VideoStatuses.Finished
			}, nil
		},
	}
	h := NewHandler(ms, ms, &mockTranscoder{})

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.GET("/videos/:id/stream/:file", h.GetVideoStream)

	url := fmt.Sprintf("/videos/%s/stream/playlist.m3u8", id)
	c.Request, _ = http.NewRequest("GET", url, nil)
	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d; want 400", w.Code)
	}
}

func TestGetVideoStream_StatusOk(t *testing.T) {
	gin.SetMode(gin.TestMode)

	id := uuid.New()
	ms := &mockService{
		findVideoById: func(ctx context.Context, id pgtype.UUID) (repo.Video, error) {
			return repo.Video{}, nil
		},
		getFile: func(ctx context.Context, key string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("fake content")), nil
		},
	}
	h := NewHandler(ms, ms, &mockTranscoder{})

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.GET("/videos/:id/stream/:file", h.GetVideoStream)

	url := fmt.Sprintf("/videos/%s/stream/playlist.m3u8", id)
	c.Request, _ = http.NewRequest("GET", url, nil)
	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusOK {
		t.Errorf("got %d; want 200", w.Code)
	}
}
