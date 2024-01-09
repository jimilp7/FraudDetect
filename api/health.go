package api

import "github.com/gin-gonic/gin"

// healthCheck godoc
// @Summary Health check
// @Description A simple health check endpoint to verify if the API is up and running.
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string "message"
// @Router /health [get]
func healthCheck(c *gin.Context) {
	// Simple response to indicate the API is working
	c.JSON(200, gin.H{
		"message": "API is up and running",
	})
}
