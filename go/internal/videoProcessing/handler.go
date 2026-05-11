package videoProcessing

import (
	"fmt"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filenameParts := strings.Split(file.Filename, ".")
	if len(filenameParts) < 1 {
		json.WriteError(c, http.StatusBadRequest, "unable to find file extension")
		return
	}

	fileExtension := filenameParts[len(filenameParts)-1]
	invalidFileType := fileExtension != "webm" && fileExtension != "mp4"
	if invalidFileType {
		json.WriteError(c, http.StatusBadRequest, "supported file formats are webm or mp4")
		return
	}

	// TODO Trancode using FFMPEG

	// TODO Save to COS

	dst := filepath.Join("./files/", filepath.Base("new-file-name."), fileExtension)
	err = c.SaveUploadedFile(file, dst)
	if err != nil {
		json.WriteError(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "file uploaded"})
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
