package storage

import (
	"fmt"
	"net/http"

	"github.com/IkBenJur/streaming-service/internal/json"
	repo "github.com/IkBenJur/streaming-service/internal/postgres/sqlc"
	"github.com/gin-gonic/gin"
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

// type UploadRequest struct {
//       Title       string `json:"title"        binding:"required"`
//       Description string `json:"description"  binding:"required,max=500"`
//       IsPublic    bool   `json:"is_public"`
//   }

//   func (h *Handler) UploadVideo(c *gin.Context) {
//       var req UploadRequest
//       if err := c.ShouldBindJSON(&req); err != nil {
//           c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//           return
//       }

func (h *Handler) CreateVideoAndGetUploadUrl(c *gin.Context) {
	id, err := h.Querier.CreateVideo(c.Request.Context(), repo.CreateVideoParams{
		Status:        repo.VideoStatuses.Pending,
		FileExtension: "mp4",
	})
	if err != nil {
		json.WriteError(c, http.StatusInternalServerError, "failed to create video entry")
		return
	}

	key := GetRawKey(fmt.Sprintf("%x.%s", id.Bytes, "mp4"))
	presignedUrl, err := h.s3Client.GenerateRawUploadUrl(c, key)
	if err != nil {
		json.WriteError(c, http.StatusInternalServerError, "failed to create url")
		return
	}

	json.WriteJSON(c, http.StatusCreated, gin.H{
		"id":         id,
		"upload-url": presignedUrl,
	})
}
