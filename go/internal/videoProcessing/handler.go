package videoProcessing

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
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
	fileFormPart, err := getFilePartFromRequest(c.Request)
	if err != nil {
		json.WriteError(c, http.StatusBadRequest, err.Error())
		return
	}

	fileExtension, err := validateFileAndGetFileExtension(fileFormPart)
	if err != nil {
		json.WriteError(c, http.StatusBadRequest, err.Error())
		return
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
	if err = h.service.SaveFileToRawStorage(fileFormPart, filename); err != nil {
		json.WriteError(c, http.StatusInternalServerError, "failed to save video")
		return
	}

	h.transcoder.Submit(id)

	c.JSON(http.StatusOK, gin.H{"message": "file uploaded"})
}

func getFilePartFromRequest(r *http.Request) (*multipart.Part, error) {
	mr, err := r.MultipartReader()
	if err != nil {
		return nil, err
	}

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			return nil, fmt.Errorf("unable to find file formpart")
		}
		if err != nil {
			return nil, err
		}

		if part.FormName() == "file" {
			return part, nil
		}
	}
}

func validateFileAndGetFileExtension(part *multipart.Part) (string, error) {
	filenameParts := strings.Split(part.FileName(), ".")
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
