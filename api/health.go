package api

import "github.com/gin-gonic/gin"

func healthCheck(c *gin.Context) {
	// Simple response to indicate the API is working
	c.JSON(200, gin.H{
		"message": "API is up and running",
	})
}
