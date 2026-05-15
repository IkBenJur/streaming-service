package videoProcessing

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/IkBenJur/streaming-service/internal/json"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service VideoProcessingService
}

func NewHandler(service VideoProcessingService) *Handler {
	return &Handler{
		service: service,
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

	// TODO Create video entry

	filename := fmt.Sprintf("new-file-name.%s", fileExtension)
	h.service.SaveFileToRawStorage(c, file, filename)

	// TODO Trancode using FFMPEG

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
