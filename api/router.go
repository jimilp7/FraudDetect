package api

import (
	"github.com/gin-gonic/gin"
)

// Function to initialize routes
func InitRoutes(router *gin.Engine) {
	router.POST("/upload", uploadTransactions)
	router.POST("/analyze/:fileID", analyzeTransactions)
	router.GET("/analyze/:analysisId/status", checkAnalysisStatus)
	router.GET("/results", getAnalysisResults)
	router.GET("/health", healthCheck)
}
