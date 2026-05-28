package videos

import (
	"context"
	"net/http"

	"github.com/IkBenJur/streaming-service/internal/json"
	repo "github.com/IkBenJur/streaming-service/internal/postgres/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service interface {
	FindVideoById(ctx context.Context, id pgtype.UUID) (repo.Video, error)
	ListVideos(ctx context.Context) ([]repo.Video, error)
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) FindById(c *gin.Context) {
	video, ok := VideoFromContext(c)
	if !ok {
		json.WriteErrorFromString(c, http.StatusInternalServerError, "video not found in context")
		return
	}

	json.WriteJSON(c, http.StatusOK, video)
}

func (h *Handler) ListVideos(c *gin.Context) {
	videos, err := h.service.ListVideos(c)
	if err != nil {
		json.WriteErrorFromStringWithErrorObjectLog(c, http.StatusInternalServerError, "not found", err)
		return
	}

	json.WriteJSON(c, http.StatusOK, videos)
}

func (h *Handler) VideoStatusIsFinished(c *gin.Context) {
	video, ok := VideoFromContext(c)
	if !ok {
		json.WriteErrorFromString(c, http.StatusInternalServerError, "video not found in context")
		return
	}

	json.WriteJSON(c, http.StatusOK, gin.H{"video_is_finished": video.Status == repo.VideoStatuses.Finished})
}
