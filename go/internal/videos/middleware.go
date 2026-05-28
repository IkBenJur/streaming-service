package videos

import (
	"net/http"

	"github.com/IkBenJur/streaming-service/internal/json"
	repo "github.com/IkBenJur/streaming-service/internal/postgres/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type contextKey string

const VideoKey contextKey = "video"

func RequireVideo(service Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		parsed, err := uuid.Parse(c.Param("id"))
		if err != nil {
			json.WriteErrorFromString(c, http.StatusBadRequest, "invalid id")
			c.Abort()
			return
		}

		id := pgtype.UUID{Bytes: parsed, Valid: true}
		video, err := service.FindVideoById(c, id)
		if err != nil {
			json.WriteErrorFromStringWithErrorObjectLog(c, http.StatusNotFound, "video not found", err)
			c.Abort()
			return
		}

		c.Set(string(VideoKey), video)
		c.Next()
	}
}

func VideoFromContext(c *gin.Context) (repo.Video, bool) {
	val, exists := c.Get(string(VideoKey))
	if !exists {
		return repo.Video{}, false
	}
	video, ok := val.(repo.Video)
	return video, ok
}
