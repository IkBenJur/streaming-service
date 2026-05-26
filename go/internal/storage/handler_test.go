package storage

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	repo "github.com/IkBenJur/streaming-service/internal/postgres/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockQuerier struct {
	repo.Querier
	createVideo                             func(ctx context.Context, arg repo.CreateVideoParams) (pgtype.UUID, error)
	videoHasValidStatusToStartProcessingJob func(ctx context.Context, id pgtype.UUID) (bool, error)
	findVideoById                           func(ctx context.Context, id pgtype.UUID) (repo.Video, error)
}

func (m *mockQuerier) CreateVideo(ctx context.Context, arg repo.CreateVideoParams) (pgtype.UUID, error) {
	return m.createVideo(ctx, arg)
}

func (m *mockQuerier) VideoHasValidStatusToStartProcessingJob(ctx context.Context, id pgtype.UUID) (bool, error) {
	return m.videoHasValidStatusToStartProcessingJob(ctx, id)
}

func (m *mockQuerier) FindVideoById(ctx context.Context, id pgtype.UUID) (repo.Video, error) {
	return m.findVideoById(ctx, id)
}

type mockStorageClient struct {
	generateRawUploadUrl    func(ctx context.Context, key string) (string, error)
	generatePresignedGetURL func(ctx context.Context, key string) (string, error)
	fileExists              func(ctx context.Context, key string) (bool, error)
	getRawKey               func(filename string) string
}

func (m *mockStorageClient) GenerateRawUploadUrl(ctx context.Context, key string) (string, error) {
	return m.generateRawUploadUrl(ctx, key)
}

func (m *mockStorageClient) GeneratePresignedGetURL(ctx context.Context, key string) (string, error) {
	return m.generatePresignedGetURL(ctx, key)
}

func (m *mockStorageClient) FileExists(ctx context.Context, key string) (bool, error) {
	return m.fileExists(ctx, key)
}

func (m *mockStorageClient) GetRawKey(filename string) string {
	return m.getRawKey(filename)
}

type mockTranscoder struct {
	submitted []pgtype.UUID
}

func (m *mockTranscoder) Submit(id pgtype.UUID) {
	m.submitted = append(m.submitted, id)
}

func TestCreateVideoAndGetUploadUrl_InvalidFileExtension(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(&mockQuerier{}, &mockStorageClient{}, &mockTranscoder{})

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.POST("/videos/create-and-get-upload-url", h.CreateVideoAndGetUploadUrl)

	body := bytes.NewBufferString(`{"title":"test","file_name":"test.txt"}`)
	c.Request, _ = http.NewRequest("POST", "/videos/create-and-get-upload-url", body)
	c.Request.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d; want 400", w.Code)
	}
}

func TestCreateVideoAndGetUploadUrl_StatusCreated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	id := uuid.New()
	videoId := pgtype.UUID{Bytes: [16]byte(id), Valid: true}

	mq := &mockQuerier{
		createVideo: func(ctx context.Context, arg repo.CreateVideoParams) (pgtype.UUID, error) {
			return videoId, nil
		},
	}
	ms := &mockStorageClient{
		getRawKey: func(filename string) string {
			return fmt.Sprintf("raw/%s", filename)
		},
		generateRawUploadUrl: func(ctx context.Context, key string) (string, error) {
			return "https://example.com/upload", nil
		},
	}
	h := NewHandler(mq, ms, &mockTranscoder{})

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.POST("/videos/create-and-get-upload-url", h.CreateVideoAndGetUploadUrl)

	body := bytes.NewBufferString(`{"title":"test","file_name":"test.mp4"}`)
	c.Request, _ = http.NewRequest("POST", "/videos/create-and-get-upload-url", body)
	c.Request.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusCreated {
		t.Errorf("got %d; want 201", w.Code)
	}
	if !strings.Contains(w.Body.String(), "upload-url") {
		t.Errorf("response missing upload-url field: %s", w.Body.String())
	}
}

func TestSubmitVideoProcessJob_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(&mockQuerier{}, &mockStorageClient{}, &mockTranscoder{})

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.POST("/videos/:id/process", h.SubmitVideoProcessJob)

	c.Request, _ = http.NewRequest("POST", "/videos/not-a-uuid/process", nil)
	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d; want 400", w.Code)
	}
}

func TestSubmitVideoProcessJob_VideoNotValidState(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mq := &mockQuerier{
		videoHasValidStatusToStartProcessingJob: func(ctx context.Context, id pgtype.UUID) (bool, error) {
			return false, nil
		},
	}
	h := NewHandler(mq, &mockStorageClient{}, &mockTranscoder{})

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.POST("/videos/:id/process", h.SubmitVideoProcessJob)

	url := fmt.Sprintf("/videos/%s/process", uuid.New())
	c.Request, _ = http.NewRequest("POST", url, nil)
	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusConflict {
		t.Errorf("got %d; want 409", w.Code)
	}
}

func TestSubmitVideoProcessJob_FileNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	id := uuid.New()
	videoId := pgtype.UUID{Bytes: [16]byte(id), Valid: true}

	mq := &mockQuerier{
		videoHasValidStatusToStartProcessingJob: func(ctx context.Context, vid pgtype.UUID) (bool, error) {
			return true, nil
		},
		findVideoById: func(ctx context.Context, vid pgtype.UUID) (repo.Video, error) {
			return repo.Video{ID: videoId, FileExtension: "mp4"}, nil
		},
	}
	ms := &mockStorageClient{
		getRawKey: func(filename string) string {
			return fmt.Sprintf("raw/%s", filename)
		},
		fileExists: func(ctx context.Context, key string) (bool, error) {
			return false, nil
		},
	}
	h := NewHandler(mq, ms, &mockTranscoder{})

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.POST("/videos/:id/process", h.SubmitVideoProcessJob)

	url := fmt.Sprintf("/videos/%s/process", id)
	c.Request, _ = http.NewRequest("POST", url, nil)
	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusNotFound {
		t.Errorf("got %d; want 404", w.Code)
	}
}

func TestSubmitVideoProcessJob_StatusOk(t *testing.T) {
	gin.SetMode(gin.TestMode)

	id := uuid.New()
	videoId := pgtype.UUID{Bytes: [16]byte(id), Valid: true}
	mt := &mockTranscoder{}

	mq := &mockQuerier{
		videoHasValidStatusToStartProcessingJob: func(ctx context.Context, vid pgtype.UUID) (bool, error) {
			return true, nil
		},
		findVideoById: func(ctx context.Context, vid pgtype.UUID) (repo.Video, error) {
			return repo.Video{ID: videoId, FileExtension: "mp4"}, nil
		},
	}
	ms := &mockStorageClient{
		getRawKey: func(filename string) string {
			return fmt.Sprintf("raw/%s", filename)
		},
		fileExists: func(ctx context.Context, key string) (bool, error) {
			return true, nil
		},
	}
	h := NewHandler(mq, ms, mt)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.POST("/videos/:id/process", h.SubmitVideoProcessJob)

	url := fmt.Sprintf("/videos/%s/process", id)
	c.Request, _ = http.NewRequest("POST", url, nil)
	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusOK {
		t.Errorf("got %d; want 200", w.Code)
	}
	if len(mt.submitted) != 1 {
		t.Errorf("expected 1 transcoder submission, got %d", len(mt.submitted))
	}
}

func TestUploadRawFile_StatusOk(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tmpDir := t.TempDir()
	localStore := NewLocalStorage(tmpDir, "http://localhost:8080")
	h := NewLocalUploadHandler(localStore)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.PUT("/videos/upload-raw/:filename", h.UploadRawFile)

	body := strings.NewReader("fake video content")
	c.Request, _ = http.NewRequest("PUT", "/videos/upload-raw/test.mp4", body)
	r.ServeHTTP(w, c.Request)

	if w.Code != http.StatusOK {
		t.Errorf("got %d; want 200", w.Code)
	}
}
