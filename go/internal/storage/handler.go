package storage

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/IkBenJur/streaming-service/internal/json"
	repo "github.com/IkBenJur/streaming-service/internal/postgres/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Handler struct {
	repo.Querier
	s3Client S3Storage
}

func NewHandler(querier repo.Querier, s3Client S3Storage) *Handler {
	return &Handler{
		Querier:  querier,
		s3Client: s3Client,
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

	key := GetRawKey(fmt.Sprintf("%x.%s", id.Bytes, fileExentension))
	presignedUrl, err := h.s3Client.GenerateRawUploadUrl(c, key)
	if err != nil {
		json.WriteErrorFromString(c, http.StatusInternalServerError, "failed to create url")
		return
	}

	json.WriteJSON(c, http.StatusCreated, gin.H{
		"id":         id,
		"upload-url": presignedUrl,
	})
}

type SubmitVideoProcessJobRequest struct {
	ID string `json:"id" binding:"required"`
}

func (h *Handler) SubmitVideoProcessJob(c *gin.Context) {
	var req SubmitVideoProcessJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		json.WriteError(c, http.StatusBadRequest, err)
		return
	}

	parsed, err := uuid.Parse(req.ID)
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

	// TODO Validate files exists in COS

	// TODO Get files from COS to local disk

	// TODO Submit transcode job

	json.WriteSucces(c, http.StatusOK, "processing job submitted")
}
