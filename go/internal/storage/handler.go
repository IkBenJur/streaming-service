package storage

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/IkBenJur/streaming-service/internal/json"
	repo "github.com/IkBenJur/streaming-service/internal/postgres/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type StorageClient interface {
	GenerateRawUploadUrl(ctx context.Context, key string) (string, error)
	FileExists(ctx context.Context, key string) (bool, error)
	GetRawKey(filename string) string
}

type Transcoder interface {
	Submit(id pgtype.UUID)
}

type Handler struct {
	repo.Querier
	storageClient StorageClient
	transcoder    Transcoder
}

func NewHandler(querier repo.Querier, storageClient StorageClient, transcoder Transcoder) *Handler {
	return &Handler{
		Querier:       querier,
		storageClient: storageClient,
		transcoder:    transcoder,
	}
}

type CreateVideoAndGetUploadUrlRequest struct {
	Title    string `json:"title" binding:"required"`
	FileName string `json:"file_name" binding:"required"`
}

func (req *CreateVideoAndGetUploadUrlRequest) validateFileAndGetExtension() (string, error) {
	filenameParts := strings.Split(req.FileName, ".")
	if len(filenameParts) < 2 {
		return "", fmt.Errorf("unable to find file extension")
	}

	fileExtension := filenameParts[len(filenameParts)-1]
	invalidFileType := fileExtension != "webm" && fileExtension != "mp4"
	if invalidFileType {
		return "", fmt.Errorf("supported file formats are webm or mp4")
	}

	return fileExtension, nil
}

func (h *Handler) CreateVideoAndGetUploadUrl(c *gin.Context) {
	var req CreateVideoAndGetUploadUrlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		json.WriteError(c, http.StatusBadRequest, err)
		return
	}

	fileExentension, err := req.validateFileAndGetExtension()
	if err != nil {
		json.WriteError(c, http.StatusBadRequest, err)
		return
	}

	id, err := h.Querier.CreateVideo(c.Request.Context(), repo.CreateVideoParams{
		Status:        repo.VideoStatuses.Pending,
		FileExtension: fileExentension,
	})
	if err != nil {
		json.WriteErrorFromString(c, http.StatusInternalServerError, "failed to create video entry")
		return
	}

	key := h.storageClient.GetRawKey(fmt.Sprintf("%x.%s", id.Bytes, fileExentension))
	presignedUrl, err := h.storageClient.GenerateRawUploadUrl(c, key)
	if err != nil {
		json.WriteErrorFromString(c, http.StatusInternalServerError, "failed to create url")
		return
	}

	json.WriteJSON(c, http.StatusCreated, gin.H{
		"id":         id,
		"upload-url": presignedUrl,
	})
}

func (h *Handler) SubmitVideoProcessJob(c *gin.Context) {
	parsed, err := uuid.Parse(c.Param("id"))
	if err != nil {
		json.WriteErrorFromString(c, http.StatusBadRequest, "invalid id format")
		return
	}
	id := pgtype.UUID{Bytes: parsed, Valid: true}

	valid, err := h.Querier.VideoHasValidStatusToStartProcessingJob(c.Request.Context(), id)
	if err != nil {
		json.WriteErrorFromString(c, http.StatusInternalServerError, "failed to validate video status")
		return
	}
	if !valid {
		json.WriteErrorFromString(c, http.StatusConflict, "video is not in a valid state to start processing")
		return
	}

	video, err := h.Querier.FindVideoById(c, id)
	if err != nil {
		json.WriteErrorFromString(c, http.StatusNotFound, "failed to find video")
		return
	}

	key := h.storageClient.GetRawKey(fmt.Sprintf("%x.%s", id.Bytes, video.FileExtension))
	fileExists, err := h.storageClient.FileExists(c, key)
	if err != nil {
		json.WriteErrorFromStringWithErrorObjectLog(c, http.StatusInternalServerError, "failed to find file", err)
		return
	}

	if !fileExists {
		json.WriteErrorFromString(c, http.StatusNotFound, "failed to find file")
		return
	}

	h.transcoder.Submit(video.ID)

	json.WriteSucces(c, http.StatusOK, "processing job submitted")
}

type LocalUploadHandler struct {
	localStore *LocalStorage
}

func NewLocalUploadHandler(localStore *LocalStorage) *LocalUploadHandler {
	return &LocalUploadHandler{localStore: localStore}
}

func (h *LocalUploadHandler) UploadRawFile(c *gin.Context) {
	filename := c.Param("filename")
	if err := h.localStore.SaveFileToRawStorage(c.Request.Body, filename); err != nil {
		json.WriteErrorFromString(c, http.StatusInternalServerError, "failed to save file")
		return
	}
	c.Status(http.StatusOK)
}
