package json

import "github.com/gin-gonic/gin"

func WriteError(c *gin.Context, status int, error string) {
	c.JSON(status, gin.H{"error": error})
}

func WriteSucces(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"message": message})
}
