package videoProcessing

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/IkBenJur/streaming-service/internal/json"
	repo "github.com/IkBenJur/streaming-service/internal/postgres/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type Transcoder interface {
	Submit(id pgtype.UUID)
}

type Handler struct {
	service    VideoProcessingService
	transcoder Transcoder
}

func NewHandler(service VideoProcessingService, transcoder Transcoder) *Handler {
	return &Handler{
		service:    service,
		transcoder: transcoder,
	}
}

func (h *Handler) UploadVideo(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		json.WriteError(c, http.StatusBadRequest, err.Error())
		return
	}

	fileExtension, err := validateFileAndGetFileExtension(file)
	if err != nil {
		json.WriteError(c, http.StatusBadRequest, err.Error())
	}

	id, err := h.service.CreateVideo(c.Request.Context(), repo.CreateVideoParams{
		Status:        repo.VideoStatuses.Pending,
		FileExtension: fileExtension,
	})
	if err != nil {
		json.WriteError(c, http.StatusInternalServerError, "failed to create video entry")
		return
	}

	filename := fmt.Sprintf("%x.%s", id.Bytes, fileExtension)
	h.service.SaveFileToRawStorage(c, file, filename)

	h.transcoder.Submit(id)

	c.JSON(http.StatusOK, gin.H{"message": "file uploaded"})
}

func validateFileAndGetFileExtension(file *multipart.FileHeader) (string, error) {
	filenameParts := strings.Split(file.Filename, ".")
	if len(filenameParts) < 1 {
		return "", fmt.Errorf("unable to find file extension")
	}

	fileExtension := filenameParts[len(filenameParts)-1]
	invalidFileType := fileExtension != "webm" && fileExtension != "mp4"
	if invalidFileType {
		return "", fmt.Errorf("supported file formats are webm or mp4")
	}

	return fileExtension, nil
}

func (h *Handler) StreamVideo(c *gin.Context) {
	fileName := c.Param("fileName")

	// Sanatize name
	fileName = filepath.Base(fileName)

	start, err := strconv.ParseInt(c.Query("start"), 10, 64)
	if err != nil {
		json.WriteError(c, http.StatusBadRequest, "invalid start")
		return
	}

	end, err := strconv.ParseInt(c.Query("end"), 10, 64)
	if err != nil {
		json.WriteError(c, http.StatusBadRequest, "invalid end")
		return
	}

	bytes, err := h.service.GetChunk(fileName, start, end)
	if err != nil {
		json.WriteError(c, http.StatusInternalServerError, "failed to get file")
		return
	}

	c.Header("Content-Type", "video/webm")
	c.Header("Content-Range", fmt.Sprintf("bytes %d-%d", start, end))
	c.Status(http.StatusPartialContent)
	c.Writer.Write(bytes)

}
