package videoProcessing

import (
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/IkBenJur/streaming-service/internal/json"
	repo "github.com/IkBenJur/streaming-service/internal/postgres/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

func (h *Handler) GetVideoStream(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		json.WriteErrorFromString(c, http.StatusBadRequest, "invalid id")
		return
	}

	videoId := pgtype.UUID{Bytes: [16]byte(id), Valid: true}
	video, err := h.service.FindVideoById(c, videoId)
	if err != nil {
		json.WriteErrorFromString(c, http.StatusNotFound, "invalid id")
		return
	}

	if video.Status != repo.VideoStatuses.Finished {
		json.WriteErrorFromString(c, http.StatusBadRequest, "video status not set to finished")
		return
	}

	file := c.Param("file")
	if strings.ContainsAny(file, "/\\") {
		json.WriteErrorFromString(c, http.StatusBadRequest, "invalid file")
		return
	}

	key := filepath.Join("hls", fmt.Sprintf("%x", videoId.Bytes), file)
	slog.Info(key)
	rc, err := h.service.GetFile(c.Request.Context(), key)
	if err != nil {
		json.WriteErrorFromString(c, http.StatusNotFound, "file not found")
		return
	}
	defer rc.Close()

	contentType := "video/mp4"
	if file == "playlist.m3u8" {
		contentType = "application/vnd.apple.mpegurl"
	}

	c.DataFromReader(http.StatusOK, -1, contentType, rc, nil)
}

func (h *Handler) UploadVideo(c *gin.Context) {
	fileFormPart, err := getFilePartFromRequest(c.Request)
	if err != nil {
		json.WriteErrorFromString(c, http.StatusBadRequest, err.Error())
		return
	}

	fileExtension, err := validateFileAndGetFileExtension(fileFormPart)
	if err != nil {
		json.WriteErrorFromString(c, http.StatusBadRequest, err.Error())
		return
	}

	id, err := h.service.CreateVideo(c.Request.Context(), repo.CreateVideoParams{
		Status:        repo.VideoStatuses.Pending,
		FileExtension: fileExtension,
	})
	if err != nil {
		json.WriteErrorFromString(c, http.StatusInternalServerError, "failed to create video entry")
		return
	}

	filename := fmt.Sprintf("%x.%s", id.Bytes, fileExtension)
	if err = h.service.SaveFileToRawStorage(fileFormPart, filename); err != nil {
		json.WriteErrorFromString(c, http.StatusInternalServerError, "failed to save video")
		return
	}

	h.transcoder.Submit(id)

	c.JSON(http.StatusCreated, gin.H{"id": id})
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
