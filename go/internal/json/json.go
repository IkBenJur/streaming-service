package json

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

func WriteError(c *gin.Context, status int, error string) {
	slog.Error(error)
	c.JSON(status, gin.H{"error": error})
}

func WriteSucces(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"message": message})
}
