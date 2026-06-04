package storage

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/IkBenJur/streaming-service/internal/json"
	repo "github.com/IkBenJur/streaming-service/internal/postgres/sqlc"
	"github.com/IkBenJur/streaming-service/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type StorageClient interface {
	GenerateRawUploadUrl(ctx context.Context, key string) (string, error)
	GeneratePresignedGetURL(ctx context.Context, key string) (string, error)
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
	id, err := utils.ParseUUID(c.Param("id"))
	if err != nil {
		json.WriteErrorFromString(c, http.StatusBadRequest, "invalid id format")
		return
	}

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

	// Update status to processing
	h.Querier.UpdateVideoStatus(c, repo.UpdateVideoStatusParams{
		ID:     id,
		Status: repo.VideoStatuses.Processing,
	})

	h.transcoder.Submit(video.ID)

	json.WriteSucces(c, http.StatusOK, "processing job submitted")
}

func (h *Handler) GetSegmentSignedUrl(c *gin.Context) {
	id, err := utils.ParseUUID(c.Param("id"))
	if err != nil {
		json.WriteErrorFromString(c, http.StatusBadRequest, "invalid id format")
		return
	}

	video, err := h.Querier.FindVideoById(c, id)
	if err != nil {
		json.WriteErrorFromString(c, http.StatusNotFound, "video not found")
		return
	}
	if video.Status != repo.VideoStatuses.Finished {
		json.WriteErrorFromString(c, http.StatusBadRequest, "video is not finished")
		return
	}

	file := c.Param("file")
	if strings.ContainsAny(file, "/\\") {
		json.WriteErrorFromString(c, http.StatusBadRequest, "invalid file")
		return
	}

	key := fmt.Sprintf("hls/%x/%s", id.Bytes, file)
	fileExists, err := h.storageClient.FileExists(c, key)
	if err != nil {
		json.WriteErrorFromStringWithErrorObjectLog(c, http.StatusInternalServerError, "failed to find file", err)
		return
	}

	if !fileExists {
		json.WriteErrorFromString(c, http.StatusNotFound, "failed to find file")
		return
	}

	signedUrl, err := h.storageClient.GeneratePresignedGetURL(c, key)
	if err != nil {
		json.WriteErrorFromString(c, http.StatusInternalServerError, "failed to generate signed url")
		return
	}

	json.WriteJSON(c, http.StatusOK, gin.H{"signed_url": signedUrl})
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

func (h *LocalUploadHandler) ServeHlsFile(c *gin.Context) {
	id := c.Param("id")
	file := c.Param("file")

	if strings.ContainsAny(id, "/\\") || strings.ContainsAny(file, "/\\") {
		c.Status(http.StatusBadRequest)
		return
	}

	key := fmt.Sprintf("hls/%s/%s", id, file)
	rc, err := h.localStore.GetFile(c.Request.Context(), key)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	defer rc.Close()

	contentType := "video/mp4"
	if strings.HasSuffix(file, ".m3u8") {
		contentType = "application/vnd.apple.mpegurl"
	}

	c.DataFromReader(http.StatusOK, -1, contentType, rc, nil)
}
