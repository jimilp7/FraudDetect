package api

import (
	"FraudDetection/services"
	"github.com/gin-gonic/gin"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"net/http"
)

type analyzeRequest struct {
	Rules []string `json:"rules"`
}

func analyzeTransactions(c *gin.Context) {
	// Extract file ID from params
	FileID := c.Param("fileID")
	var req analyzeRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Extract rules
	rules := req.Rules

	// Generate a unique analysis ID
	analysisID, err := gonanoid.New()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate an analysis ID"})
		return
	}

	// Create a new engine instance
	engine := services.NewFraudDetectionEngine(FileID, analysisID, rules)

	// Store the engine instance
	EnginesMutex.Lock()
	Engines[analysisID] = engine
	EnginesMutex.Unlock()

	// Start analysis in a goroutine
	go engine.RunEngine()

	// Respond with the analysis ID
	c.JSON(http.StatusOK, gin.H{"analysisId": analysisID})
}

func checkAnalysisStatus(c *gin.Context) {
	analysisID := c.Param("analysisId")

	// Retrieve the engine instance
	EnginesMutex.Lock()
	engine, exists := Engines[analysisID]
	EnginesMutex.Unlock()

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
