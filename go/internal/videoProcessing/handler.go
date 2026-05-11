package videoProcessing

import (
	"net/http"
	"os"
	"path/filepath"
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

	filePath := filepath.Join("./files/", fileName)
	file, err := os.Open(filePath)
	if err != nil {
		json.WriteError(c, http.StatusNotFound, "video not found")
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		json.WriteError(c, http.StatusInternalServerError, "could not stat file")
	}

	http.ServeContent(c.Writer, c.Request, stat.Name(), stat.ModTime(), file)
}
