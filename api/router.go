package api

import (
	docs "FraudDetection/docs"
	"FraudDetection/services"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"sync"
)

// Spin up Engines as needed for the async analyze endpoint
var (
	Engines      = make(map[string]*services.FraudDetectionEngine)
	EnginesMutex = &sync.Mutex{}
)

func InitRoutes(router *gin.Engine) {
	docs.SwaggerInfo.BasePath = "/api/v1"
	router.POST("/upload", uploadTransactions)
	router.POST("/analyze/:fileID", analyzeTransactions)
	router.GET("/analyze/:analysisId/status", checkAnalysisStatus)
	router.GET("/results/:analysisId", getAnalysisResults)
	router.GET("/health", healthCheck)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
