package videoProcessing

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
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
	findVideoById        func(ctx context.Context, id pgtype.UUID) (repo.Video, error)
	createVideo          func(ctx context.Context, arg repo.CreateVideoParams) (pgtype.UUID, error)
	saveFileToRawStorage func(r io.Reader, filename string) error
	getFile              func(ctx context.Context, key string) (io.ReadCloser, error)
	rawFilePath          func(filename string) string
	hlsOutputPath        func(id string) (string, error)
}

func (m *mockService) FindVideoById(ctx context.Context, id pgtype.UUID) (repo.Video, error) {
	return m.findVideoById(ctx, id)
}

func (m *mockService) CreateVideo(ctx context.Context, arg repo.CreateVideoParams) (pgtype.UUID, error) {
	return m.createVideo(ctx, arg)
}

func (m *mockService) SaveFileToRawStorage(r io.Reader, filename string) error {
	return m.saveFileToRawStorage(r, filename)
}

func (m *mockService) GetFile(ctx context.Context, key string) (io.ReadCloser, error) {
	return m.getFile(ctx, key)
}

func (m *mockService) RawFilePath(filename string) string {
	return m.rawFilePath(filename)
}

func (m *mockService) HLSOutputPath(id string) (string, error) {
	return m.hlsOutputPath(id)
}

type mockTranscoder struct {
	submitted []pgtype.UUID
}

func (m *mockTranscoder) Submit(id pgtype.UUID) {
	m.submitted = append(m.submitted, id)
}

func TestUploadVideo_ValidateFileExtension(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ms := &mockService{}
	h := NewHandler(ms, ms, &mockTranscoder{})

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.POST("/videos/upload", h.UploadVideo)

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, _ := mw.CreateFormFile("file", "test.txt")
	part.Write([]byte("fake video bytes"))
	mw.Close()

	c.Request, _ = http.NewRequest("POST", "/videos/upload", &body)
	c.Request.Header.Set("Content-Type", mw.FormDataContentType())

	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d; want 400", w.Code)
	}
}

func TestUploadVideo_StatusOk(t *testing.T) {

	id := uuid.New()
	videoId := pgtype.UUID{Bytes: [16]byte(id), Valid: true}
	gin.SetMode(gin.TestMode)
	ms := &mockService{
		createVideo: func(ctx context.Context, arg repo.CreateVideoParams) (pgtype.UUID, error) {
			return videoId, nil
		},
		saveFileToRawStorage: func(r io.Reader, filename string) error {
			return nil
		},
	}
	h := NewHandler(ms, ms, &mockTranscoder{})

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.POST("/videos/upload", h.UploadVideo)

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, _ := mw.CreateFormFile("file", "test.mp4")
	part.Write([]byte("fake video bytes"))
	mw.Close()

	c.Request, _ = http.NewRequest("POST", "/videos/upload", &body)
	c.Request.Header.Set("Content-Type", mw.FormDataContentType())

	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusCreated {
		t.Errorf("got %d; want 201", w.Code)
	}
}

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
