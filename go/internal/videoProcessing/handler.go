package videoProcessing

import (
	"net/http"
	"path/filepath"

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

	// TODO Reject wrong file types

	// TODO Trancode using FFMPEG

	// TODO Save to COS

	dst := filepath.Join("./files/", filepath.Base("new-file-name.txt"))
	err = c.SaveUploadedFile(file, dst)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "file uploaded"})
}
