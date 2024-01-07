package api

import (
	"FraudDetection/services"
	"github.com/gin-gonic/gin"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"net/http"
	"sync"
)

var (
	engines      = make(map[string]*services.FraudDetectionEngine)
	enginesMutex = &sync.Mutex{}
)

type AnalysisStatus struct {
	Status   string // "processing", "complete", "failed"
	ResultID string // Initially empty, updated upon completion
}

func analyzeTransactions(c *gin.Context) {
	// Extract file ID from params
	FileID := c.Param("fileID")

	// Generate a unique analysis ID
	analysisID, err := gonanoid.New()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate an analysis ID"})
		return
	}

	// Create a new engine instance
	engine := services.NewFraudDetectionEngine(FileID, analysisID)

	// Store the engine instance
	enginesMutex.Lock()
	engines[analysisID] = engine
	enginesMutex.Unlock()

	// Start analysis in a goroutine
	go engine.RunAnalysis()

	// Respond with the analysis ID
	c.JSON(http.StatusOK, gin.H{"analysisId": analysisID})
}

//func simulateAnalysis(analysisID, fileID string) {
//}

func checkAnalysisStatus(c *gin.Context) {
	analysisID := c.Param("analysisId")

	// Retrieve the engine instance
	enginesMutex.Lock()
	engine, exists := engines[analysisID]
	enginesMutex.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Analysis not found"})
		return
	}

	// Get the status from the engine's analysisMap
	engine.AnalysisLock.Lock()
	status, exists := engine.AnalysisMap[analysisID]
	engine.AnalysisLock.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Status not found"})
		return
	}

	// Respond with the status
	c.JSON(http.StatusOK, status)
}
