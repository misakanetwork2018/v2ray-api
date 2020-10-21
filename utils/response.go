package utils

import "github.com/gin-gonic/gin"

func RespondWithError(code int, message string, c *gin.Context) {
	c.JSON(code, gin.H{
		"success": false,
		"msg":     message,
	})
	c.Abort()
}
