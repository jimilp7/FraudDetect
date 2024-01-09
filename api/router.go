package api

import (
	"FraudDetection/services"
	"github.com/gin-gonic/gin"
	"sync"
)

// Spin up Engines as needed for the async analyze endpoint
var (
	Engines      = make(map[string]*services.FraudDetectionEngine)
	EnginesMutex = &sync.Mutex{}
)

func InitRoutes(router *gin.Engine) {
	router.POST("/upload", uploadTransactions)
	router.POST("/analyze/:fileID", analyzeTransactions)
	router.GET("/analyze/:analysisId/status", checkAnalysisStatus)
	router.GET("/results/:analysisId", getAnalysisResults)
	router.GET("/health", healthCheck)
}
